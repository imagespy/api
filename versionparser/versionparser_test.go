package versionparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	testcases := []struct {
		version             string
		expectedDistinction string
		expectedString      string
		testName            string
	}{
		{"1", "major", "1", "Major as integer"},
		{"1-alpine", "major-alpine", "1-alpine", "Major with build suffix"},
		{"v1", "major", "v1", "Major with v prefix"},
		{"v1-alpine", "major-alpine", "v1-alpine", "Major with v prefix and build suffix"},
		{"1.2", "majorMinor", "1.2", "MajorMinor bare"},
		{"1.2-alpine", "majorMinor-alpine", "1.2-alpine", "MajorMinor with build suffix"},
		{"v1.2", "majorMinor", "v1.2", "MajorMinor with v prefix"},
		{"v1.2-alpine", "majorMinor-alpine", "v1.2-alpine", "MajorMinor with v prefix and build suffix"},
		{"1.2.3", "majorMinorPatch", "1.2.3", "MajorMinorPatch bare"},
		{"1.2.3-alpine", "majorMinorPatch-alpine", "1.2.3-alpine", "MajorMinorPatch with build suffix"},
		{"v1.2.3", "majorMinorPatch", "v1.2.3", "MajorMinorPatch with v prefix"},
		{"v1.2.3-alpine", "majorMinorPatch-alpine", "v1.2.3-alpine", "MajorMinorPatch with v prefix and build suffix"},
		{"ubuntu-20180913", "nameDate-ubuntu", "ubuntu-20180913", "NameDate"},
		{"latest", "static-latest", "latest", "Static latest"},
		{"mainline", "static-mainline", "mainline", "Static mainline"},
		{"master", "static-master", "master", "Static master"},
		{"stable", "static-stable", "stable", "Static stable"},
		{"sometag", "unknown-sometag", "sometag", "Unknown"},
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			vp := Registry.FindForVersion(tc.version)
			assert.Equal(t, tc.expectedDistinction, vp.Distinction())
			assert.Equal(t, tc.expectedString, vp.String())
		})
	}
}
