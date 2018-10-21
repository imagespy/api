package store

import (
	"fmt"
	"time"
)

type Model struct {
	ID int
}

type Feature struct {
	Model
	CreatedAt time.Time
	Name      string
}

func (Feature) TableName() string {
	return "imagespy_feature"
}

type Image struct {
	Model
	CreatedAt     time.Time
	Digest        string
	Name          string
	Platforms     []*Platform
	SchemaVersion int
	ScrapedAt     time.Time
	Tags          []*Tag
}

func (Image) TableName() string {
	return "imagespy_image"
}

func (i *Image) HasTag(t *Tag) bool {
	for _, tag := range i.Tags {
		if tag.Name == t.Name {
			return true
		}
	}

	return false
}

func (i *Image) FindTag(name string) (*Tag, error) {
	for _, t := range i.Tags {
		if t.Name == name {
			return t, nil
		}
	}

	return nil, fmt.Errorf("image %s does not have tag %s", i.Name, name)
}

func (i *Image) FindLatestTagByDistiction(d string) (*Tag, error) {
	for _, t := range i.Tags {
		if t.IsLatest && t.Distinction == d {
			return t, nil
		}
	}

	return nil, fmt.Errorf("image %s does not have tag with distinction", i.Name, d)
}

type Layer struct {
	Model
	Digest       string
	SourceImages []*Image `gorm:"many2many:imagespy_layer_source_images;"`
}

func (Layer) TableName() string {
	return "imagespy_layer"
}

type LayerOfPlatform struct {
	Model
	Layer      *Layer
	LayerID    int
	Platform   *Platform
	PlatformID int
	Position   int
}

func (LayerOfPlatform) TableName() string {
	return "imagespy_layerofplatform"
}

type OSFeature struct {
	Model
	CreatedAt time.Time
	Name      string
}

func (OSFeature) TableName() string {
	return "imagespy_osfeature"
}

type Platform struct {
	Model
	Architecture   string
	Created        time.Time
	CreatedAt      time.Time
	Features       []*Feature `gorm:"many2many:imagespy_platform_features;"`
	ImageID        int
	Layers         []*Layer
	ManifestDigest string
	OS             string
	OSFeatures     []*OSFeature `gorm:"many2many:imagespy_platform_os_features;"`
	OSVersion      string
	Variant        string
}

func (Platform) TableName() string {
	return "imagespy_platform"
}

type Tag struct {
	Model
	Distinction string
	ImageID     int
	IsLatest    bool
	IsTagged    bool
	Name        string
}

func (Tag) TableName() string {
	return "imagespy_tag"
}
