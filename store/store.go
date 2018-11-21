package store

import (
	"errors"
)

var (
	// ErrDoesNotExist is returned if the requested model could not be found in the store.
	ErrDoesNotExist = errors.New("Model does not exist")
)

// Store represents the high-level API to access models.
type Store interface {
	Images() ImageStore
	Layers() LayerStore
	LayerPositions() LayerPositionStore
	Platforms() PlatformStore
	Tags() TagStore
}

// ImageStore allows creating, manipulating and reading images.
type ImageStore interface {
	Create(i *Image) error
	Get(o ImageGetOptions) (*Image, error)
	List(o ImageListOptions) ([]*Image, error)
	Update(i *Image) error
}

// ImageGetOptions is used to query image with certain criterias.
type ImageGetOptions struct {
	Digest         string
	ID             int
	Name           string
	TagDistinction string
	TagIsLatest    *bool
	TagName        string
}

type ImageListOptions struct {
	Digest string
	Name   string
}

type LayerStore interface {
	Create(l *Layer) error
	Get(o LayerGetOptions) (*Layer, error)
	List(o LayerListOptions) ([]*Layer, error)
	Update(l *Layer) error
}

type LayerGetOptions struct {
	Digest string
	ID     int
}

type LayerListOptions struct {
	PlatformID int
}

type LayerPositionStore interface {
	Create(*LayerPosition) error
	List(o LayerPositionListOptions) ([]*LayerPosition, error)
}

type LayerPositionListOptions struct {
	LayerID    int
	PlatformID int
}

type PlatformStore interface {
	Create(*Platform) error
	Get(o PlatformGetOptions) (*Platform, error)
	List(o PlatformListOptions) ([]*Platform, error)
}

type PlatformGetOptions struct {
	Architecture string
	ImageID      int
	OS           string
	OSVersion    string
	Variant      string
}

type PlatformListOptions struct {
	LayerDigest string
}

type TagStore interface {
	Create(*Tag) error
	Get(o TagGetOptions) (*Tag, error)
	List(o TagListOptions) ([]*Tag, error)
	Update(*Tag) error
}

type TagGetOptions struct {
	Distinction string
	ImageID     int
	ImageName   string
	IsLatest    *bool
	Name        string
}

type TagListOptions struct {
	ImageID int
}
