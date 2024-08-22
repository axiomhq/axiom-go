// The purpose of this example is to show how to replicate the contents of
// Hacker News into Axiom.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

const (
	baseURL    = "https://hacker-news.firebaseio.com"
	maxWorkers = 100
)

var httpClient = axiom.DefaultHTTPClient()

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	dataset := os.Getenv("AXIOM_DATASET")
	if dataset == "" {
		log.Fatal("AXIOM_DATASET is required")
	}

	// 1. Get the ID of the latest item. This is where ingestion will stop.
	maxItemID, err := getMaxItemID()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Generate a channel of IDs that will be used to fetch the items.
	idCh := generateIDs(maxItemID)

	// 3. Fetch the items and create a channel of events to consume for
	// ingestion.
	eventCh := fetchEvents(idCh)

	// 4. Wrap the channel that streams up to maxItemID events to present a
	// progress bar.
	bar := progressbar.NewOptions64(int64(maxItemID),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetItsString("events"),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionThrottle(time.Millisecond*65),
	)
	progressEventCh := make(chan axiom.Event)
	go func() {
		for event := range eventCh {
			progressEventCh <- event
			_ = bar.Add(1)
		}
		close(progressEventCh)
		if finishErr := bar.Finish(); finishErr != nil {
			log.Fatal(finishErr)
		}
	}()

	// 5. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 6. Ingest ⚡
	res, err := client.IngestChannel(ctx, dataset, progressEventCh,
		// Have the server use the "time" field as the event timestamp.
		ingest.SetTimestampField("time"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 7. Make sure everything went smoothly.
	//
	// Note: If you ever make it here, you have ingested all of Hacknews into
	// Axiom. Congratulations... I guess?! 🤔
	for _, fail := range res.Failures {
		log.Print(fail.Error)
	}
}

func getMaxItemID() (uint64, error) {
	res, err := httpClient.Get(baseURL + "/v0/maxitem.json")
	if err != nil {
		return 0, fmt.Errorf("failed to get maxitem.json: %w", err)
	}
	defer res.Body.Close()

	maxItemIDBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read body: %w", err)
	}

	maxItemID, err := strconv.ParseUint(string(maxItemIDBytes), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse maxitem.json: %w", err)
	}

	return maxItemID, nil
}

func generateIDs(max uint64) <-chan uint64 {
	ch := make(chan uint64, maxWorkers*10)
	go func() {
		for i := range max + 1 {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

func fetchEvents(eventIDs <-chan uint64) <-chan axiom.Event {
	var (
		eventCh        = make(chan axiom.Event, 1024)
		workerErrGroup errgroup.Group
	)

	go func() {
		if err := workerErrGroup.Wait(); err != nil {
			log.Fatal(err)
		}
		close(eventCh)
	}()

	workerErrGroup.SetLimit(maxWorkers)

	for range maxWorkers {
		workerErrGroup.Go(func() error {
			for id := range eventIDs {
				resp, err := httpClient.Get(fmt.Sprintf("%s/v0/item/%d.json", baseURL, id))
				if err != nil {
					return err
				}

				var event axiom.Event
				if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
					_ = resp.Body.Close()
					return err
				} else if err = resp.Body.Close(); err != nil {
					return err
				}

				eventCh <- event
			}
			return nil
		})
	}

	return eventCh
}
