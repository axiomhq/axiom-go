package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAPIToken(t *testing.T) {
	assert.True(t, IsAPIToken(apiToken))
	assert.False(t, IsAPIToken(personalToken))
	assert.False(t, IsAPIToken(unspecifiedToken))
}

func TestIsPersonalToken(t *testing.T) {
	assert.False(t, IsPersonalToken(apiToken))
	assert.True(t, IsPersonalToken(personalToken))
	assert.False(t, IsPersonalToken(unspecifiedToken))
}

func TestIsValidToken(t *testing.T) {
	assert.True(t, IsValidToken(apiToken))
	assert.True(t, IsValidToken(personalToken))
	assert.False(t, IsValidToken(unspecifiedToken))
}
