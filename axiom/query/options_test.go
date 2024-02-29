package query_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/axiom-go/axiom/query"
)

func TestOptions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		options []query.Option
		want    query.Options
	}{
		{
			name: "set start time",
			options: []query.Option{
				query.SetStartTime(now.String()),
			},
			want: query.Options{
				StartTime: now.String(),
			},
		},
		{
			name: "set end time",
			options: []query.Option{
				query.SetEndTime(now.String()),
			},
			want: query.Options{
				EndTime: now.String(),
			},
		},
		{
			name: "set start time (apl)",
			options: []query.Option{
				query.SetStartTime("now"),
			},
			want: query.Options{
				StartTime: "now",
			},
		},
		{
			name: "set end time (apl)",
			options: []query.Option{
				query.SetEndTime("now"),
			},
			want: query.Options{
				EndTime: "now",
			},
		},
		{
			name: "set curser include",
			options: []query.Option{
				query.SetCursor("123", true),
			},
			want: query.Options{
				Cursor:        "123",
				IncludeCursor: true,
			},
		},
		{
			name: "set curser exclude",
			options: []query.Option{
				query.SetCursor("123", false),
			},
			want: query.Options{
				Cursor:        "123",
				IncludeCursor: false,
			},
		},
		{
			name: "set event label",
			options: []query.Option{
				query.SetVariable("foo", "bar"),
			},
			want: query.Options{
				Variables: map[string]any{
					"foo": "bar",
				},
			},
		},
		{
			name: "set multiple event labels",
			options: []query.Option{
				query.SetVariable("foo", "bar"),
				query.SetVariable("bar", "foo"),
			},
			want: query.Options{
				Variables: map[string]any{
					"foo": "bar",
					"bar": "foo",
				},
			},
		},
		{
			name: "set event labels",
			options: []query.Option{
				query.SetVariables(map[string]any{
					"foo": "bar",
					"bar": "foo",
				}),
			},
			want: query.Options{
				Variables: map[string]any{
					"foo": "bar",
					"bar": "foo",
				},
			},
		},
		{
			name: "set event labels on existing labels",
			options: []query.Option{
				query.SetVariable("movie", "spider man"),
				query.SetVariables(map[string]any{
					"foo": "bar",
					"bar": "foo",
				}),
			},
			want: query.Options{
				Variables: map[string]any{
					"foo": "bar",
					"bar": "foo",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var options query.Options
			for _, option := range tt.options {
				option(&options)
			}
			assert.Equal(t, tt.want, options)
		})
	}
}
