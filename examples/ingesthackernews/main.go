// The purpose of this example is to show how to replicate the contents of
// Hacker News into Axiom.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"

	"github.com/axiomhq/axiom-go/axiom"
)

const (
	baseURL    = "https://hacker-news.firebaseio.com"
	maxWorkers = 100
)

func main() {
	// Export `AXIOM_DATASET` in addition to the required environment variables.

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
		progressbar.OptionThrottle(65*time.Millisecond),
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
	res, err := client.Datasets.IngestChannel(context.Background(), dataset, progressEventCh, axiom.IngestOptions{
		TimestampField: "time",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 7. Make sure everything went smoothly.
	//
	// Note: If you ever make it here, you have ingested all of Hacknews into
	// Axiom. Congratulations.. Or not?! 🤔
	for _, fail := range res.Failures {
		log.Print(fail.Error)
	}
}

func getMaxItemID() (uint64, error) {
	res, err := http.Get(baseURL + "/v0/maxitem.json")
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
		for i := uint64(0); i <= max; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

func fetchEvents(eventIDs <-chan uint64) <-chan axiom.Event {
	var (
		eventCh        = make(chan axiom.Event, maxWorkers*10)
		workerErrGroup errgroup.Group
	)

	go func() {
		if err := workerErrGroup.Wait(); err != nil {
			log.Fatal(err)
		}
		close(eventCh)
	}()

	workerErrGroup.SetLimit(maxWorkers)

	for i := 0; i < maxWorkers; i++ {
		workerErrGroup.Go(func() error {
			var event axiom.Event
			for id := range eventIDs {
				res, err := http.Get(fmt.Sprintf("%s/v0/item/%d.json", baseURL, id))
				if err != nil {
					return err
				}

				if err := json.NewDecoder(res.Body).Decode(&event); err != nil {
					_ = res.Body.Close()
					return err
				} else if err = res.Body.Close(); err != nil {
					return err
				}

				eventCh <- event
			}
			return nil
		})
	}

	return eventCh
}
