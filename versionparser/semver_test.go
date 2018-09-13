package versionparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMajor(t *testing.T) {
	first, _ := majorFactory("1")
	second, _ := majorFactory("2")
	result, err := first.IsGreaterThan(second)
	assert.NoError(t, err)
	assert.False(t, result)

	result, err = second.IsGreaterThan(first)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestMajorMinor(t *testing.T) {
	first, _ := majorMinorFactory("1.1")
	second, _ := majorMinorFactory("1.2")
	result, err := first.IsGreaterThan(second)
	assert.NoError(t, err)
	assert.False(t, result)

	result, err = second.IsGreaterThan(first)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestMajorMinorPatch(t *testing.T) {
	first, _ := majorMinorPatchFactory("1.2.3")
	second, _ := majorMinorPatchFactory("1.2.4")
	result, err := first.IsGreaterThan(second)
	assert.NoError(t, err)
	assert.False(t, result)

	result, err = second.IsGreaterThan(first)
	assert.NoError(t, err)
	assert.True(t, result)
}
