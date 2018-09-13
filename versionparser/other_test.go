package versionparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameDate(t *testing.T) {
	first, _ := nameDateFactory("ubuntu-20180913")
	second, _ := nameDateFactory("ubuntu-20180914")
	result, err := first.IsGreaterThan(second)
	assert.NoError(t, err)
	assert.False(t, result)

	result, err = second.IsGreaterThan(first)
	assert.NoError(t, err)
	assert.True(t, result)
}
