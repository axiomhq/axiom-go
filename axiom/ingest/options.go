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
