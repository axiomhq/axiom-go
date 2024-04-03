package ingest

// TimestampField is the default field the server will look for a timestamp to
// use as the ingestion time. If not present, the server will set the ingestion
// time to the current server time.
const TimestampField = "_time"

// Options specifies the optional parameters for ingestion.
type Options struct {
	// TimestampField defines a custom field to extract the ingestion timestamp
	// from. Defaults to [TimestampField].
	TimestampField string `url:"timestamp-field,omitempty"`
	// TimestampFormat defines a custom format for the [Options.TimestampField].
	// The reference time is "Mon Jan 2 15:04:05 -0700 MST 2006", as specified
	// in https://pkg.go.dev/time/?tab=doc#Parse.
	TimestampFormat string `url:"timestamp-format,omitempty"`
	// CSVDelimiter is the delimiter that separates CSV fields. Only valid when
	// the content to be ingested is CSV formatted.
	CSVDelimiter string `url:"csv-delimiter,omitempty"`
	// EventLabels are a key-value pairs that will be added to all events. Their
	// purpose is to allow for labeling events without alterting the original
	// event data. This is especially useful when ingesting events from a
	// third-party source that you do not have control over.
	EventLabels map[string]any `url:"-"`
	// Fields is a list of fields to be ingested with every event. This is only
	// valid for CSV content and also completely optional. It comes in handy
	// when the CSV content does not have a header row.
	CSVFields []string `url:"-"`
}

// An Option applies optional parameters to an ingest operation.
type Option func(*Options)

// SetTimestampField specifies the field Axiom will use to extract the events
// time from. Defaults to [TimestampField]
func SetTimestampField(field string) Option {
	return func(o *Options) { o.TimestampField = field }
}

// SetTimestampFormat specifies the format of the timestamp field. The reference
// time is "Mon Jan 2 15:04:05 -0700 MST 2006", as specified in
// https://pkg.go.dev/time/?tab=doc#Parse.
func SetTimestampFormat(format string) Option {
	return func(o *Options) { o.TimestampFormat = format }
}

// SetCSVDelimiter specifies the delimiter that separates CSV fields. Only valid
// when the content to be ingested is CSV formatted.
func SetCSVDelimiter(delim string) Option {
	return func(o *Options) { o.CSVDelimiter = delim }
}

// SetEventLabel adds a label to apply to all events. This option can be called
// multiple times to add multiple labels. If a label with the same key already
// exists, it will be overwritten.
func SetEventLabel(key string, value any) Option {
	return func(o *Options) {
		if o.EventLabels == nil {
			o.EventLabels = make(map[string]any, 1)
		}
		o.EventLabels[key] = value
	}
}

// SetEventLabels sets the labels to apply to all events. It will overwrite any
// existing labels.
func SetEventLabels(labels map[string]any) Option {
	return func(o *Options) { o.EventLabels = labels }
}

// AddCSVField adds one or more fields to be ingested with every CSV event.
func AddCSVField(field ...string) Option {
	return func(o *Options) {
		if o.CSVFields == nil {
			o.CSVFields = make([]string, 0, len(field))
		}
		o.CSVFields = append(o.CSVFields, field...)
	}
}

// SetCSVFields sets the fields to be ingested with every CSV event.
func SetCSVFields(fields ...string) Option {
	return func(o *Options) { o.CSVFields = fields }
}
