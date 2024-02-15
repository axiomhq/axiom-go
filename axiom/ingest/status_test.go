package ingest_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/axiom-go/axiom/ingest"
)

func TestStatus_Add(t *testing.T) {
	now := time.Now()

	s1 := ingest.Status{
		Ingested: 2,
		Failed:   1,
		Failures: []*ingest.Failure{
			{
				Timestamp: now,
				Error:     "I am an error",
			},
		},
		ProcessedBytes: 1024,
		BlocksCreated:  0,
		WALLength:      2048,
	}

	s2 := ingest.Status{
		Ingested: 3,
		Failed:   1,
		Failures: []*ingest.Failure{
			{
				Timestamp: now.Add(-time.Second),
				Error:     "I am another error",
			},
		},
		ProcessedBytes: 1024,
		BlocksCreated:  1,
		WALLength:      1024,
	}

	s1.Add(&s2)

	assert.EqualValues(t, 5, s1.Ingested)
	assert.EqualValues(t, 2, s1.Failed)
	assert.Len(t, s1.Failures, 2)
	assert.EqualValues(t, 2048, s1.ProcessedBytes)
	assert.EqualValues(t, 1, s1.BlocksCreated)
	assert.EqualValues(t, 1024, s1.WALLength)
}
