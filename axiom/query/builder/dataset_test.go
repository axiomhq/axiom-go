package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataset_String(t *testing.T) {
	tests := []struct {
		dataset Dataset
		want    string
	}{
		{"foo", "foo"},
		{"fooBar", "fooBar"},
		{"foo-bar", "['foo-bar']"},
		{"foo_bar", "['foo_bar']"},
		{"foo.bar", "['foo.bar']"},
	}
	for _, tt := range tests {
		t.Run(string(tt.dataset), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.dataset.String())
		})
	}
}
