package store

import (
	"errors"
)

var (
	ErrDoesNotExist = errors.New("Model does not exist")
)

type Store interface {
	// CreateImageFromRegistryImage(distinction string, regImg registry.Image) (*Image, *Tag, error)
	// FindLatestTagWithImage(repository, tag string) (*Tag, error)
	// FindLatestImageWithTagsByDistinction(distinction string, repository string) (*Image, error)
	// FindImageWithTagsByTag(repository string, tag string) (*Image, error)
	// UpdateTag(*Tag) error
	Images() ImageStore
	Tags() TagStore
}

type ImageStore interface {
	Create(i *Image) error
	Get(o ImageGetOptions) (*Image, error)
	List(o ImageListOptions) ([]*Image, error)
	Update(i *Image) error
}

type ImageGetOptions struct {
	Digest string
	ID     int
}

type ImageListOptions struct {
	Digest         string
	Name           string
	TagDistinction string
	TagIsLatest    *bool
	TagName        string
}

type TagStore interface {
	Get(o TagGetOptions) (*Tag, error)
	Update(*Tag) error
}

type TagGetOptions struct {
	Distinction string
	ImageID     int
	ImageName   string
	IsLatest    *bool
	Name        string
}
