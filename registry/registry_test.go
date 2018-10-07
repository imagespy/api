package registry

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageV1(t *testing.T) {
	resultLines := []string{}

	img, err := NewImage("quay.io/prometheus/prometheus:v2.3.2", Opts{Insecure: false})
	assert.NoError(t, err)
	digest, err := img.Digest()
	assert.NoError(t, err)
	schemaVersion, err := img.SchemaVersion()
	assert.NoError(t, err)
	tag, err := img.Tag()
	assert.NoError(t, err)

	resultLines = append(resultLines, fmt.Sprintf("%s - %s - %d", tag, digest, schemaVersion))

	platforms, err := img.Platforms()
	assert.NoError(t, err)
	for _, p := range platforms {
		resultLine := fmt.Sprintf("%s - %s - %s - %s - %s - %s - %s", p.Digest(), strings.Join(p.Features(), ","), strings.Join(p.OSFeatures(), ","), p.OSVersion(), p.Variant(), p.Architecture(), p.OS())
		resultLines = append(resultLines, resultLine)
		manifest, err := p.Manifest()
		assert.NoError(t, err)
		resultLines = append(resultLines, fmt.Sprintf("%s - %d", manifest.MediaType(), manifest.SchemaVersion()))
		for _, l := range manifest.Layers() {
			digest, err := l.Digest()
			assert.NoError(t, err)
			_, err = l.MediaType()
			assert.Error(t, err)
			_, err = l.Size()
			assert.Error(t, err)
			resultLines = append(resultLines, digest)
		}

		config, err := manifest.Config()
		assert.NoError(t, err)
		_, err = config.MediaType()
		assert.Error(t, err)
		_, err = config.Size()
		assert.Error(t, err)
		resultLines = append(resultLines, config.Digest().String())
		history, err := config.History()
		assert.NoError(t, err)
		for _, h := range history {
			resultLines = append(resultLines, h.Created.String())
		}
	}

	resultLines = append(resultLines, "")
	fixtureFile := "./fixtures/registry_test/quay_io_prometheus_prometheus_v2.3.2.txt"
	b, err := ioutil.ReadFile(fixtureFile)
	if err != nil {
		assert.Failf(t, "Error reading fixture file", "Error reading fixture file %s", fixtureFile)
	}

	fixture := string(b[:])
	assert.Equal(t, strings.Split(fixture, "\n"), resultLines)
}

func TestImageV2(t *testing.T) {
	testcases := []struct {
		fixture string
		image   string
		name    string
	}{
		{fixture: "python_3.6.6.txt", image: "python:3.6.6", name: "Multiple Platforms"},
		{fixture: "prom_prometheus_v2.3.2.txt", image: "prom/prometheus:v2.3.2", name: "Single Platform"},
	}

	for _, tc := range testcases {
		resultLines := []string{}

		img, err := NewImage(tc.image, Opts{Insecure: false})
		assert.NoError(t, err)
		digest, err := img.Digest()
		assert.NoError(t, err)
		schemaVersion, err := img.SchemaVersion()
		assert.NoError(t, err)
		tag, err := img.Tag()
		assert.NoError(t, err)

		resultLines = append(resultLines, fmt.Sprintf("%s - %s - %d", tag, digest, schemaVersion))

		platforms, err := img.Platforms()
		assert.NoError(t, err)
		for _, p := range platforms {
			resultLine := fmt.Sprintf("%s - %s - %s - %s - %s - %s", p.Digest(), strings.Join(p.Features(), ","), strings.Join(p.OSFeatures(), ","), p.Variant(), p.Architecture(), p.OS())
			resultLines = append(resultLines, resultLine)
			manifest, err := p.Manifest()
			assert.NoError(t, err)
			resultLines = append(resultLines, fmt.Sprintf("%s - %d", manifest.MediaType(), manifest.SchemaVersion()))
			for _, l := range manifest.Layers() {
				digest, err := l.Digest()
				assert.NoError(t, err)
				mediaType, err := l.MediaType()
				assert.NoError(t, err)
				size, err := l.Size()
				assert.NoError(t, err)
				resultLines = append(resultLines, fmt.Sprintf("%s - %s - %d", digest, mediaType, size))
			}

			config, err := manifest.Config()
			assert.NoError(t, err)
			mediaType, err := config.MediaType()
			assert.NoError(t, err)
			size, err := config.Size()
			assert.NoError(t, err)
			resultLines = append(resultLines, fmt.Sprintf("%s - %s - %d", config.Digest(), mediaType, size))
			history, err := config.History()
			assert.NoError(t, err)
			for _, h := range history {
				resultLines = append(resultLines, h.Created.String())
			}
		}

		resultLines = append(resultLines, "")
		fixtureFile := "./fixtures/registry_test/" + tc.fixture
		b, err := ioutil.ReadFile(fixtureFile)
		if err != nil {
			assert.Failf(t, "Error reading fixture file", "Error reading fixture file %s", fixtureFile)
		}

		fixture := string(b[:])
		assert.Equal(t, strings.Split(fixture, "\n"), resultLines)
	}
}
