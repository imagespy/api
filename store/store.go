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
}

type ImageStore interface {
	Create(i *Image) error
	Get(digest string) (*Image, error)
	List(o ImageListOptions) ([]*Image, error)
	Update(i *Image) error
}

type ImageListOptions struct {
	Digest         string
	Name           string
	TagDistinction string
	TagIsLatest    *bool
	TagName        string
}
