package store

import (
	"errors"

	"github.com/imagespy/api/registry"
)

var (
	ErrDoesNotExist = errors.New("Model does not exist")
)

type Store interface {
	Close() error
	CreateImageFromRegistryImage(distinction string, regImg *registry.Image) (*Image, *Tag, error)
	FindImageByTag(repository string, tag string) (*Image, error)
	Migrate() error
}
