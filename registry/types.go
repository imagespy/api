package registry

import (
	"github.com/docker/docker/image"
	digest "github.com/opencontainers/go-digest"
)

type Config interface {
	Digest() digest.Digest
	History() ([]image.History, error)
	MediaType() (string, error)
	Size() (int, error)
}

type Layer interface {
	Digest() (string, error)
	MediaType() (string, error)
	Size() (int, error)
}

type Manifest interface {
	Config() (Config, error)
	Layers() []Layer
	MediaType() string
	SchemaVersion() int
}

type Platform interface {
	Architecture() string
	Digest() digest.Digest
	Features() []string
	Manifest() (Manifest, error)
	OS() string
	OSFeatures() []string
	OSVersion() string
	Variant() string
}
