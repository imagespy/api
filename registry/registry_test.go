package registry

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

type resultImage struct {
	Digest        string            `json:"digest"`
	Platforms     []*resultPlatform `json:"platforms"`
	SchemaVersion int               `json:"schema_version"`
	Tag           string            `json:"tag"`
}

type resultPlatform struct {
	Architecture string          `json:"architecture"`
	Digest       string          `json:"digest"`
	Features     []string        `json:"features"`
	Manifest     *resultManifest `json:"manifest"`
	OS           string          `json:"os"`
	OSFeatures   []string        `json:"os_features"`
	OSVersion    string          `json:"os_version"`
	Variant      string          `json:"variant"`
}

type resultManifest struct {
	Config        *resultManifestConfig  `json:"config"`
	Layers        []*resultManifestLayer `json:"layers"`
	MediaType     string                 `json:"media_type"`
	SchemaVersion int                    `json:"schema_version"`
}

type resultManifestConfig struct {
	Digest    string   `json:"digest"`
	History   []string `json:"history"`
	MediaType string   `json:"media_type"`
	Size      int      `json:"size"`
}

type resultManifestLayer struct {
	Digest    string `json:"digest"`
	MediaType string `json:"media_type"`
	Size      int    `json:"size"`
}

func TestImageV2(t *testing.T) {
	testcases := []struct {
		fixture string
		image   string
		name    string
	}{
		{fixture: "python_3.6.6.json", image: "python:3.6.6", name: "Multiple Platforms"},
		{fixture: "prom_prometheus_v2.3.2.json", image: "prom/prometheus:v2.3.2", name: "Single Platform"},
		{fixture: "quay_io_prometheus_prometheus_v2.3.2.json", image: "quay.io/prometheus/prometheus:v2.3.2", name: "Single Platform V1"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			r := &resultImage{}
			img, err := NewImage(tc.image, Opts{Insecure: false})
			assert.NoError(t, err)
			digest, err := img.Digest()
			r.Digest = digest
			assert.NoError(t, err)
			schemaVersion, err := img.SchemaVersion()
			assert.NoError(t, err)
			r.SchemaVersion = schemaVersion
			tag, err := img.Tag()
			assert.NoError(t, err)
			r.Tag = tag

			platforms, err := img.Platforms()
			assert.NoError(t, err)
			for _, p := range platforms {
				rp := &resultPlatform{
					Architecture: p.Architecture(),
					Digest:       p.Digest().String(),
					Features:     p.Features(),
					OS:           p.OS(),
					OSVersion:    p.OSVersion(),
					OSFeatures:   p.OSFeatures(),
					Variant:      p.Variant(),
				}
				manifest, err := p.Manifest()
				assert.NoError(t, err)
				rm := &resultManifest{
					MediaType:     manifest.MediaType(),
					SchemaVersion: manifest.SchemaVersion(),
				}
				for _, l := range manifest.Layers() {
					rml := &resultManifestLayer{}
					digest, err := l.Digest()
					assert.NoError(t, err)
					rml.Digest = digest
					rml.MediaType = l.MediaType()
					rml.Size = l.Size()
					rm.Layers = append(rm.Layers, rml)
				}

				rmc := &resultManifestConfig{}
				config, err := manifest.Config()
				assert.NoError(t, err)
				rmc.Digest = config.Digest().String()
				rmc.MediaType = config.MediaType()
				rmc.Size = config.Size()
				history, err := config.History()
				assert.NoError(t, err)
				for _, h := range history {
					rmc.History = append(rmc.History, h.Created.Format("2006-01-02 15:04:05 -0700 UTC"))
				}

				rm.Config = rmc
				rp.Manifest = rm
				r.Platforms = append(r.Platforms, rp)
			}

			fixtureFile := "./fixtures/registry_test/" + tc.fixture
			b, err := ioutil.ReadFile(fixtureFile)
			require.NoErrorf(t, err, "Error reading fixture file %s", fixtureFile)

			fixture := &resultImage{}
			err = json.Unmarshal(b, fixture)
			require.NoError(t, err)
			assert.EqualValues(t, fixture, r)
		})
	}
}
