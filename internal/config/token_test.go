package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAPIToken(t *testing.T) {
	assert.True(t, IsAPIToken(apiToken))
	assert.False(t, IsAPIToken(personalToken))
	assert.False(t, IsAPIToken(sharedAccessSignature))
	assert.False(t, IsAPIToken(unspecifiedToken))
}

func TestIsPersonalToken(t *testing.T) {
	assert.False(t, IsPersonalToken(apiToken))
	assert.True(t, IsPersonalToken(personalToken))
	assert.False(t, IsPersonalToken(sharedAccessSignature))
	assert.False(t, IsPersonalToken(unspecifiedToken))
}

func TestIsSharedAccessSignature(t *testing.T) {
	assert.False(t, IsSharedAccessSignature(apiToken))
	assert.False(t, IsSharedAccessSignature(personalToken))
	assert.True(t, IsSharedAccessSignature(sharedAccessSignature))
	assert.False(t, IsSharedAccessSignature(unspecifiedToken))
}

func TestIsValidToken(t *testing.T) {
	assert.True(t, IsValidCredential(apiToken))
	assert.True(t, IsValidCredential(personalToken))
	assert.True(t, IsValidCredential(sharedAccessSignature))
	assert.False(t, IsValidCredential(unspecifiedToken))
}
