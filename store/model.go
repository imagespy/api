package store

import "time"

type Feature struct {
	CreatedAt time.Time
	ID        int
	Name      string
}

func (Feature) TableName() string {
	return "imagespy_feature"
}

type Image struct {
	CreatedAt     time.Time
	Digest        string
	ID            int
	Name          string
	Platforms     []*Platform
	SchemaVersion int
	ScrapedAt     time.Time
	Tags          []*Tag
}

func (Image) TableName() string {
	return "imagespy_image"
}

type Layer struct {
	Digest       string
	ID           int
	SourceImages []*Image `gorm:"many2many:imagespy_layer_source_images;"`
}

func (Layer) TableName() string {
	return "imagespy_layer"
}

type LayerOfPlatform struct {
	ID         int
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
	CreatedAt time.Time
	ID        int
	Name      string
}

func (OSFeature) TableName() string {
	return "imagespy_osfeature"
}

type Platform struct {
	Architecture   string
	Created        time.Time
	CreatedAt      time.Time
	Features       []*Feature `gorm:"many2many:imagespy_platform_features;"`
	Image          *Image
	ImageID        int
	ID             int
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
	Distinction string
	ImageID     int
	ID          int
	IsLatest    bool
	IsTagged    bool
	Name        string
}

func (Tag) TableName() string {
	return "imagespy_tag"
}
