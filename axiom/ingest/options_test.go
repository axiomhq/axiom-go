package ingest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/axiom-go/axiom/ingest"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []ingest.Option
		want    ingest.Options
	}{
		{
			name: "set timestamp field",
			options: []ingest.Option{
				ingest.SetTimestampField("ts"),
			},
			want: ingest.Options{
				TimestampField: "ts",
			},
		},
		{
			name: "set timestamp format",
			options: []ingest.Option{
				ingest.SetTimestampFormat("unixnano"),
			},
			want: ingest.Options{
				TimestampFormat: "unixnano",
			},
		},
		{
			name: "set csv delimiter",
			options: []ingest.Option{
				ingest.SetCSVDelimiter(";"),
			},
			want: ingest.Options{
				CSVDelimiter: ";",
			},
		},
		{
			name: "set event label",
			options: []ingest.Option{
				ingest.SetEventLabel("foo", "bar"),
			},
			want: ingest.Options{
				EventLabels: map[string]any{
					"foo": "bar",
				},
			},
		},
		{
			name: "set multiple event labels",
			options: []ingest.Option{
				ingest.SetEventLabel("foo", "bar"),
				ingest.SetEventLabel("bar", "foo"),
			},
			want: ingest.Options{
				EventLabels: map[string]any{
					"foo": "bar",
					"bar": "foo",
				},
			},
		},
		{
			name: "set event labels",
			options: []ingest.Option{
				ingest.SetEventLabels(map[string]any{
					"foo": "bar",
					"bar": "foo",
				}),
			},
			want: ingest.Options{
				EventLabels: map[string]any{
					"foo": "bar",
					"bar": "foo",
				},
			},
		},
		{
			name: "set event labels on existing labels",
			options: []ingest.Option{
				ingest.SetEventLabel("movie", "spider man"),
				ingest.SetEventLabels(map[string]any{
					"foo": "bar",
					"bar": "foo",
				}),
			},
			want: ingest.Options{
				EventLabels: map[string]any{
					"foo": "bar",
					"bar": "foo",
				},
			},
		},
		{
			name: "add csv field",
			options: []ingest.Option{
				ingest.AddCSVField("foo"),
			},
			want: ingest.Options{
				CSVFields: []string{"foo"},
			},
		},
		{
			name: "add multiple csv fields",
			options: []ingest.Option{
				ingest.AddCSVField("foo"),
				ingest.AddCSVField("bar", "baz"),
			},
			want: ingest.Options{
				CSVFields: []string{"foo", "bar", "baz"},
			},
		},
		{
			name: "set csv fields",
			options: []ingest.Option{
				ingest.SetCSVFields("foo", "bar"),
			},
			want: ingest.Options{
				CSVFields: []string{"foo", "bar"},
			},
		},
		{
			name: "set csv fields on existing csv fields",
			options: []ingest.Option{
				ingest.SetCSVFields("foo", "bar"),
				ingest.SetCSVFields("bar", "foo"),
			},
			want: ingest.Options{
				CSVFields: []string{"bar", "foo"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var options ingest.Options
			for _, option := range tt.options {
				option(&options)
			}
			assert.Equal(t, tt.want, options)
		})
	}
}
