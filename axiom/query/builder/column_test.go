package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumn_String(t *testing.T) {
	tests := []struct {
		column Column
		want   string
	}{
		{"foo", "foo"},
		{"foo_bar", "foo_bar"},
		{"foo.bar", "['foo.bar']"},
	}
	for _, tt := range tests {
		t.Run(string(tt.column), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.column.String())
		})
	}
}
