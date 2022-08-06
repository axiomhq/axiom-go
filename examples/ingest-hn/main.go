package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/axiomhq/axiom-go/axiom"
)

const BaseURL = "https://hacker-news.firebaseio.com"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
	)
	defer cancel()

	datasetName := os.Getenv("DATASET_NAME")
	if datasetName == "" {
		log.Fatal("Missing DATASET_NAME")
	}

	res, err := http.Get(BaseURL + "/v0/maxitem.json")
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get maxitem.json: %w", err))
	}
	maxItemIDBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()
		log.Fatal(fmt.Errorf("failed to read body: %w", err))
	}
	res.Body.Close()
	maxItemID, err := strconv.ParseUint(string(maxItemIDBytes), 10, 64)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse maxitem.json: %w", err))
	}

	workerWaitGroup := sync.WaitGroup{}
	workQueue := make(chan uint64, 100)
	eventChan := make(chan axiom.Event, 1000)

	go func() {
		for i := uint64(0); i <= maxItemID; i++ {
			select {
			case <-ctx.Done():
				log.Fatal(ctx.Err())
			case workQueue <- i:
			}
		}

		close(workQueue)
		workerWaitGroup.Wait()
		close(eventChan)
	}()

	for i := 0; i < 100; i++ {
		workerWaitGroup.Add(1)
		go func() {
			defer workerWaitGroup.Done()
			for id := range workQueue {
				select {
				case <-ctx.Done():
					log.Fatal(ctx.Err())
				default:
				}

				log.Printf("Fetching item %d\n", id)
				res, err := http.Get(fmt.Sprintf("%s/v0/item/%d.json", BaseURL, id))
				if err != nil {
					log.Fatal(fmt.Errorf("http request failed: %w", err))
				}

				event := make(axiom.Event)
				if err := json.NewDecoder(res.Body).Decode(&event); err != nil {
					res.Body.Close()
					log.Fatal(fmt.Errorf("failed to decode json: %w", err))
				}

				res.Body.Close()
				eventChan <- event
			}
		}()
	}

	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to construct Axiom client: %w", err))
	}

	_, err = client.Datasets.IngestChannel(ctx, datasetName, eventChan, axiom.IngestOptions{})
	if err != nil {
		log.Fatal(fmt.Errorf("failed to ingest events: %w", err))
	}
}
