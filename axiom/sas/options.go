package sas

import (
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/axiomhq/axiom-go/axiom/query"
)

// Options represents the options for creating a shared access token or a shared
// access signature.
type Options struct {
	// OrganizationID is the ID of the organization the token and signature is
	// valid for.
	OrganizationID string
	// Dataset name the token and signature is valid for.
	Dataset string
	// Filter is the top-level query filter to apply to all query requests
	// the token and signature is valid for.
	Filter query.Filter
	// MinStartTime is the earliest query start time the token and signature is
	// valid for.
	MinStartTime time.Time
	// MaxEndTime is the latest query end time the token and signature is valid
	// for.
	MaxEndTime time.Time
}

// optionsFromURLValues returns `Options` from the given `url.Values`.
func optionsFromURLValues(q url.Values) (options Options, err error) {
	options = Options{
		OrganizationID: q.Get(queryOrgID),
		Dataset:        q.Get(queryDataset),
	}

	// The filter is encoded as a JSON string.
	var (
		f         filter
		filterStr = q.Get(queryFilter)
	)
	if err = json.Unmarshal([]byte(filterStr), &f); err != nil {
		return options, err
	}
	options.Filter = f.toQueryFilter()

	if options.MinStartTime, err = time.Parse(time.RFC3339, q.Get(queryMinStartTime)); err != nil {
		return options, err
	}
	if options.MaxEndTime, err = time.Parse(time.RFC3339, q.Get(queryMaxEndTime)); err != nil {
		return options, err
	}

	return options, nil
}

// urlValues returns the options as `url.Values`.
func (o Options) urlValues() (url.Values, error) {
	// The filter is encoded as a JSON string.
	filterStr, err := json.Marshal(filterFromQueryFilter(o.Filter))
	if err != nil {
		return nil, err
	}

	q := make(url.Values, 5)
	q.Set(queryOrgID, o.OrganizationID)
	q.Set(queryDataset, o.Dataset)
	q.Set(queryFilter, string(filterStr))
	q.Set(queryMinStartTime, o.MinStartTime.Format(time.RFC3339))
	q.Set(queryMaxEndTime, o.MaxEndTime.Format(time.RFC3339))

	return q, nil
}

// validate makes sure that all options are provided.
func (o Options) validate() error {
	if o.OrganizationID == "" {
		return errors.New("organization ID is required")
	} else if o.Dataset == "" {
		return errors.New("dataset is required")
	} else if o.Filter.Op == 0 {
		return errors.New("filter is required")
	} else if o.MinStartTime.IsZero() {
		return errors.New("minimum start time is required")
	} else if o.MaxEndTime.IsZero() {
		return errors.New("maximum end time is required")
	}
	return nil
}
