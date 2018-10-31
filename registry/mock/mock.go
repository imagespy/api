package mock

import (
	"fmt"

	dockerImage "github.com/docker/docker/image"
	reg "github.com/genuinetools/reg/registry"
	"github.com/imagespy/api/registry"
	digest "github.com/opencontainers/go-digest"
)

type Image struct {
	registry.Image
	digest        string
	name          string
	platforms     []registry.Platform
	repository    *mockRepository
	schemaVersion int
	tag           string
}

func (m *Image) Digest() (string, error) {
	return m.digest, nil
}

func (m *Image) Platforms() ([]registry.Platform, error) {
	return m.platforms, nil
}

func (m *Image) Repository() registry.Repository {
	return m.repository
}

func (m *Image) SchemaVersion() (int, error) {
	return m.schemaVersion, nil
}

func (m *Image) Tag() (string, error) {
	return m.tag, nil
}

func NewImage(digest string, name string, platforms []registry.Platform, schemaVersion int, tag string) *Image {
	i := &Image{
		digest:        digest,
		name:          name,
		platforms:     platforms,
		schemaVersion: schemaVersion,
		tag:           tag,
	}
	return i
}

type mockRegistry struct {
	registry.Registry
	repositories map[string]*mockRepository
}

func (m *mockRegistry) AddImage(i *Image) {
	r, ok := m.repositories[i.name]
	if ok {
		i.repository = r
		r.images = append(r.images, i)
	} else {
		repository := &mockRepository{name: i.name, images: []registry.Image{i}}
		i.repository = repository
		m.repositories[i.name] = repository
	}
}

func (m *mockRegistry) Image(imageName string) (registry.Image, error) {
	p, err := reg.ParseImage(imageName)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%s/%s", p.Domain, p.Path)
	repository, ok := m.repositories[name]
	if !ok {
		return nil, fmt.Errorf("Unknown repository for %s", imageName)
	}

	if p.Tag != "" {
		for _, i := range repository.images {
			ci, _ := i.(*Image)
			if ci.tag == p.Tag {
				return i, nil
			}
		}
	}

	if p.Digest.String() != "" {
		for _, i := range repository.images {
			ci, _ := i.(*Image)
			if ci.digest == p.Digest.String() {
				return i, nil
			}
		}
	}

	return nil, fmt.Errorf("Unknown reference for %s", imageName)
}

func NewRegistry() *mockRegistry {
	return &mockRegistry{repositories: map[string]*mockRepository{}}
}

type mockRepository struct {
	registry.Repository
	name   string
	images []registry.Image
}

func (m *mockRepository) FullName() string {
	return m.name
}

func (m *mockRepository) Images() ([]registry.Image, error) {
	return m.images, nil
}

type mockPlatform struct {
	registry.Platform
	arch     string
	digest   string
	manifest *mockManifest
	os       string
}

func (m *mockPlatform) Architecture() string {
	return m.arch
}

func (m *mockPlatform) Digest() digest.Digest {
	return digest.Digest(m.digest)
}

func (m *mockPlatform) Features() []string {
	return []string{}
}

func (m *mockPlatform) Manifest() (registry.Manifest, error) {
	return nil, nil
}

func (m *mockPlatform) OS() string {
	return m.os
}

func (m *mockPlatform) OSFeatures() []string {
	return []string{}
}

func (m *mockPlatform) OSVersion() string {
	return ""
}

func (m *mockPlatform) Variant() string {
	return ""
}

func NewPlatform(arch string, digest string, layers []string, manifestDigest string, os string) *mockPlatform {
	return &mockPlatform{
		arch:   arch,
		digest: digest,
		manifest: &mockManifest{
			config: &mockConfig{
				digest: manifestDigest,
			},
		},
		os: os,
	}
}

type mockManifest struct {
	registry.Manifest
	config registry.Config
	layers []registry.Layer
}

func (m *mockManifest) Config() (registry.Config, error) {
	return nil, nil
}
func (m *mockManifest) Layers() []registry.Layer {
	return nil
}

type mockConfig struct {
	registry.Config
	digest  string
	history []dockerImage.History
}

func (m *mockConfig) Digest() digest.Digest {
	return digest.Digest(m.digest)
}

func (m *mockConfig) History() ([]dockerImage.History, error) {
	return m.history, nil
}

type mockLayer struct {
	registry.Layer
	digest string
}

func (m *mockLayer) Digest() (string, error) {
	return m.digest, nil
}
