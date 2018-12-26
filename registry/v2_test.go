package registry

import (
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/schema2"
)

func TestNewManifestV2(t *testing.T) {
	schema2Manifest := schema2.Manifest{
		Versioned: manifest.Versioned{
			MediaType:     "unittest",
			SchemaVersion: 2,
		},
		Layers: []distribution.Descriptor{
			{Digest: digest.FromString("abc")},
			{Digest: digest.FromString("def")},
			{Digest: digest.FromString("def")},
			{Digest: digest.FromString("ghi")},
		},
	}

	manifestV2 := NewManifestV2(schema2Manifest, &PlatformV2{})
	actualLayerDigests := []string{}
	for _, l := range manifestV2.layers {
		actualLayerDigests = append(actualLayerDigests, l.rawLayer.Digest.String())
	}
	assert.EqualValues(
		t,
		[]string{
			"sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
			"sha256:cb8379ac2098aa165029e3938a51da0bcecfc008fd6795f401178647f96c5b34",
			"sha256:50ae61e841fac4e8f9e40baf2ad36ec868922ea48368c18f9535e47db56dd7fb",
		},
		actualLayerDigests,
	)
}
