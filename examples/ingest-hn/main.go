package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/axiomhq/axiom-go/axiom"
)

const BaseURL = "https://hacker-news.firebaseio.com"

func getMaxItemID() (uint64, error) {
	res, err := http.Get(BaseURL + "/v0/maxitem.json")
	if err != nil {
		return 0, fmt.Errorf("failed to get maxitem.json: %w", err)
	}
	defer res.Body.Close()

	maxItemIDBytes, err := ioutil.ReadAll(res.Body)
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
	ch := make(chan uint64, 100)

	go func() {
		for i := uint64(0); i <= max; i++ {
			ch <- i
		}

		close(ch)
	}()

	return ch
}

func fetchEvents(eventIDs <-chan uint64) <-chan axiom.Event {
	eventChan := make(chan axiom.Event, 1000)

	workerWaitGroup := sync.WaitGroup{}
	go func() {
		workerWaitGroup.Wait()
		close(eventChan)
	}()

	for i := 0; i < 100; i++ {
		workerWaitGroup.Add(1)
		go func() {
			defer workerWaitGroup.Done()
			for id := range eventIDs {
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

	return eventChan
}

func main() {
	datasetName := os.Getenv("DATASET_NAME")
	if datasetName == "" {
		log.Fatal("Missing DATASET_NAME")
	}

	maxItemID, err := getMaxItemID()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get max item id: %w", err))
	}

	idChan := generateIDs(maxItemID)
	eventChan := fetchEvents(idChan)

	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to construct Axiom client: %w", err))
	}

	_, err = client.Datasets.IngestChannel(context.TODO(), datasetName, eventChan, axiom.IngestOptions{})
	if err != nil {
		log.Fatal(fmt.Errorf("failed to ingest events: %w", err))
	}
}
