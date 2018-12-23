package registry

import (
	dockerImage "github.com/docker/docker/image"
	digest "github.com/opencontainers/go-digest"
)

type Config interface {
	Digest() digest.Digest
	History() ([]dockerImage.History, error)
	MediaType() string
	Size() int
}

type Image interface {
	Digest() (string, error)
	Platform(arch string, os string) (Platform, error)
	Platforms() ([]Platform, error)
	Repository() Repository
	SchemaVersion() (int, error)
	Tag() (string, error)
}

type Layer interface {
	Digest() (string, error)
	MediaType() string
	Size() int
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

type Registry interface {
	Address() string
	Repository(imageName string) (Repository, error)
	Image(imageName string) (Image, error)
}

type Repository interface {
	FullName() string
	Image(digest string, tag string) Image
	Images() ([]Image, error)
}
