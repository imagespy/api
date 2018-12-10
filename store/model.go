package store

import (
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
	SchemaVersion int
	ScrapedAt     time.Time
}

func (Image) TableName() string {
	return "imagespy_image"
}

type Layer struct {
	Model
	Digest         string
	SourceImageIDs []int `gorm:"-"`
}

func (Layer) TableName() string {
	return "imagespy_layer"
}

type LayerPosition struct {
	Model
	LayerID    int
	PlatformID int
	Position   int
}

func (LayerPosition) TableName() string {
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
