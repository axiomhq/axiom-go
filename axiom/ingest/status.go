package ingest

import "time"

// Status is the status of an event ingestion operation.
type Status struct {
	// Ingested is the amount of events that have been ingested.
	Ingested uint64 `json:"ingested"`
	// Failed is the amount of events that failed to ingest.
	Failed uint64 `json:"failed"`
	// Failures are the ingestion failures, if any.
	Failures []*Failure `json:"failures"`
	// ProcessedBytes is the number of bytes processed.
	ProcessedBytes uint64 `json:"processedBytes"`
	// BlocksCreated is the amount of blocks created.
	BlocksCreated uint32 `json:"blocksCreated"`
	// WALLength is the length of the Write-Ahead Log.
	WALLength uint32 `json:"walLength"`
}

// Failure describes the ingestion failure of a single event.
type Failure struct {
	// Timestamp of the event that failed to ingest.
	Timestamp time.Time `json:"timestamp"`
	// Error that made the event fail to ingest.
	Error string `json:"error"`
}
