package registry

import (
	"github.com/docker/distribution/manifest/schema1"
	dockerImage "github.com/docker/docker/image"
	imageV1 "github.com/docker/docker/image/v1"
	reg "github.com/genuinetools/reg/registry"
	digest "github.com/opencontainers/go-digest"
)

type ConfigV1 struct {
	digest  digest.Digest
	history []dockerImage.History
}

func (c *ConfigV1) Digest() digest.Digest {
	return c.digest
}

func (c *ConfigV1) History() ([]dockerImage.History, error) {
	return c.history, nil
}

func (c *ConfigV1) MediaType() string {
	return ""
}

func (c *ConfigV1) Size() int {
	return 0
}

type PlatformV1 struct {
	digest    digest.Digest
	image     Image
	manifest  *ManifestV1
	regClient *reg.Registry
}

func (p *PlatformV1) Architecture() string {
	return "amd64"
}

func (p *PlatformV1) Digest() digest.Digest {
	return ""
}

func (p *PlatformV1) Features() []string {
	var f []string
	return f
}

func (p *PlatformV1) Manifest() (Manifest, error) {
	return p.manifest, nil
}

func (p *PlatformV1) OS() string {
	return "linux"
}

func (p *PlatformV1) OSFeatures() []string {
	var f []string
	return f
}

func (p *PlatformV1) OSVersion() string {
	return ""
}

func (p *PlatformV1) Variant() string {
	return ""
}

type ManifestV1 struct {
	config      *ConfigV1
	digest      digest.Digest
	layers      []Layer
	platform    *PlatformV1
	rawManifest *schema1.SignedManifest
}

func NewManifestV1(p *PlatformV1, m *schema1.SignedManifest, manifestDigest digest.Digest) (*ManifestV1, error) {
	layers := []Layer{}
	for _, fsLayer := range m.FSLayers {
		if fsLayer.BlobSum == digestSHA256GzippedEmptyTar {
			continue
		}

		layers = append([]Layer{&LayerV1{fsLayer.BlobSum}}, layers...)
	}

	entries := []dockerImage.History{}
	for _, entry := range m.History {
		e, err := imageV1.HistoryFromConfig([]byte(entry.V1Compatibility), false)
		if err != nil {
			return nil, err
		}

		entries = append([]dockerImage.History{e}, entries...)
	}

	return &ManifestV1{
		config: &ConfigV1{
			digest:  manifestDigest,
			history: entries,
		},
		layers:      layers,
		platform:    p,
		rawManifest: m,
	}, nil
}

func (m *ManifestV1) Config() (Config, error) {
	return m.config, nil
}

func (m *ManifestV1) Layers() []Layer {
	return m.layers
}

func (m *ManifestV1) MediaType() string {
	return m.rawManifest.MediaType
}

func (m *ManifestV1) SchemaVersion() int {
	return m.rawManifest.SchemaVersion
}

type LayerV1 struct {
	digest digest.Digest
}

func (l *LayerV1) Digest() (string, error) {
	return l.digest.String(), nil
}

func (l *LayerV1) MediaType() string {
	return ""
}

func (l *LayerV1) Size() int {
	return 0
}
