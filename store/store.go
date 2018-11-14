package store

import (
	"errors"
)

var (
	ErrDoesNotExist = errors.New("Model does not exist")
)

type Store interface {
	Images() ImageStore
	Layers() LayerStore
	Platforms() PlatformStore
	Tags() TagStore
}

type ImageStore interface {
	Create(i *Image) error
	Get(o ImageGetOptions) (*Image, error)
	List(o ImageListOptions) ([]*Image, error)
	Update(i *Image) error
}

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
	Digest     string
	PlatformID int
}

type LayerListOptions struct {
	PlatformID int
}

type PlatformStore interface {
	Create(*Platform) error
	List(o PlatformListOptions) ([]*Platform, error)
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
