package registry

import (
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	dockerImage "github.com/docker/docker/image"
	imageV1 "github.com/docker/docker/image/v1"
	reg "github.com/genuinetools/reg/registry"
	digest "github.com/opencontainers/go-digest"
)

type ConfigV2 struct {
	digest    digest.Digest
	history   []dockerImage.History
	mediaType string
	size      int
}

func (c *ConfigV2) Digest() digest.Digest {
	return c.digest
}

func (c *ConfigV2) History() ([]dockerImage.History, error) {
	return c.history, nil
}

func (c *ConfigV2) MediaType() string {
	return c.mediaType
}

func (c *ConfigV2) Size() int {
	return c.size
}

type LayerV2 struct {
	rawLayer distribution.Descriptor
}

func (l *LayerV2) Digest() (string, error) {
	return l.rawLayer.Digest.String(), nil
}

func (l *LayerV2) MediaType() string {
	return l.rawLayer.MediaType
}

func (l *LayerV2) Size() int {
	return int(l.rawLayer.Size)
}

type ManifestV2 struct {
	config        *ConfigV2
	layers        []*LayerV2
	mediaType     string
	platform      *PlatformV2
	rawManifest   schema2.Manifest
	regClient     *reg.Registry
	schemaVersion int
}

func (m *ManifestV2) Layers() []Layer {
	layers := make([]Layer, len(m.layers))
	for i, v := range m.layers {
		layers[i] = Layer(v)
	}

	return layers
}

func (m *ManifestV2) Config() (Config, error) {
	if m.config == nil {
		mV1, err := m.regClient.ManifestV1(m.platform.image.parsed.Path, m.platform.image.parsed.Tag)
		if err != nil {
			return nil, err
		}

		entries := []dockerImage.History{}
		for _, entry := range mV1.History {
			e, err := imageV1.HistoryFromConfig([]byte(entry.V1Compatibility), false)
			if err != nil {
				return nil, err
			}

			entries = append([]dockerImage.History{e}, entries...)
		}

		m.config = &ConfigV2{
			digest:    m.rawManifest.Config.Digest,
			history:   entries,
			mediaType: m.rawManifest.Config.MediaType,
			size:      int(m.rawManifest.Config.Size),
		}
	}

	return m.config, nil
}

func (m *ManifestV2) MediaType() string {
	return m.mediaType
}

func (m *ManifestV2) SchemaVersion() int {
	return m.schemaVersion
}

func NewManifestV2(m schema2.Manifest, p *PlatformV2) *ManifestV2 {
	knownDigests := map[string]struct{}{}
	layerV2s := []*LayerV2{}
	for _, rawLayer := range m.Layers {
		if rawLayer.Digest == digestSHA256GzippedEmptyTar {
			continue
		}

		digest := rawLayer.Digest.String()
		_, exists := knownDigests[digest]
		if exists {
			continue
		}

		knownDigests[digest] = struct{}{}
		layerV2s = append(layerV2s, &LayerV2{rawLayer: rawLayer})
	}

	return &ManifestV2{
		layers:        layerV2s,
		mediaType:     m.MediaType,
		platform:      p,
		rawManifest:   m,
		regClient:     p.regClient,
		schemaVersion: m.SchemaVersion,
	}
}

type PlatformV2 struct {
	architecture string
	digest       digest.Digest
	features     []string
	image        *image
	manifest     *ManifestV2
	os           string
	osFeatures   []string
	osVersion    string
	regClient    *reg.Registry
	variant      string
}

func (p *PlatformV2) Architecture() string {
	return p.architecture
}

func (p *PlatformV2) Digest() digest.Digest {
	return p.digest
}

func (p *PlatformV2) Features() []string {
	return p.features
}

func (p *PlatformV2) Manifest() (Manifest, error) {
	if p.manifest == nil {
		m, err := p.regClient.ManifestV2(p.image.parsed.Path, p.digest.String())
		if err != nil {
			return nil, err
		}

		p.manifest = NewManifestV2(m, p)
	}

	return p.manifest, nil
}

func (p *PlatformV2) OS() string {
	return p.os
}

func (p *PlatformV2) OSFeatures() []string {
	return p.osFeatures
}

func (p *PlatformV2) OSVersion() string {
	return p.osVersion
}

func (p *PlatformV2) Variant() string {
	return p.variant
}
