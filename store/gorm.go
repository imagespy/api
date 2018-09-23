package store

import (
	"time"

	"github.com/imagespy/api/registry"
	"github.com/jinzhu/gorm"
)

type gormStore struct {
	db *gorm.DB
}

func (g *gormStore) Close() error {
	return g.db.Close()
}

func (g *gormStore) FindImageByTag(repository string, tag string) (*Image, error) {
	image := &Image{}
	result := g.db.Where("imagespy_image.name = ? AND imagespy_tag.name = ? AND imagespy_tag.is_latest = 1", repository, tag).Joins("inner join imagespy_tag on imagespy_tag.image_id = imagespy_image.id").Preload("Tags").Preload("Platforms").First(image)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrDoesNotExist
		}

		return nil, result.Error
	}

	for _, p := range image.Platforms {
		features := []*Feature{}
		result := g.db.Model(p).Related(&features, "Features")
		if result.Error != nil {
			return nil, result.Error
		}

		p.Features = append(p.Features, features...)
		osFeatures := []*OSFeature{}
		result = g.db.Model(p).Related(&osFeatures, "OSFeatures")
		if result.Error != nil {
			return nil, result.Error
		}

		p.OSFeatures = append(p.OSFeatures, osFeatures...)
		layersOfPlatform := []*LayerOfPlatform{}
		result = g.db.Model(p).Related(&layersOfPlatform)
		if result.Error != nil {
			return nil, result.Error
		}

		for _, lop := range layersOfPlatform {
			layer := &Layer{}
			result := g.db.Model(lop).Related(layer)
			if result.Error != nil {
				return nil, result.Error
			}

			sourceImages := []*Image{}
			g.db.Model(layer).Related(&sourceImages)
			if result.Error != nil {
				return nil, result.Error
			}

			layer.SourceImages = sourceImages
			p.Layers = append(p.Layers, layer)
		}
	}

	return image, nil
}

func (g *gormStore) CreateImageFromRegistryImage(distinction string, regImg *registry.Image) (*Image, *Tag, error) {
	imageDigest, err := regImg.Digest()
	if err != nil {
		return nil, nil, err
	}

	imageSV, err := regImg.SchemaVersion()
	if err != nil {
		return nil, nil, err
	}

	image := &Image{
		CreatedAt:     time.Now().UTC(),
		Digest:        imageDigest,
		Name:          regImg.Repository.FullName(),
		SchemaVersion: imageSV,
		ScrapedAt:     time.Now().UTC(),
	}
	tx := g.db.Begin()
	if tx.Error != nil {
		return nil, nil, err
	}

	result := tx.FirstOrCreate(image, Image{Digest: imageDigest, Name: regImg.Repository.FullName(), SchemaVersion: imageSV})
	if result.Error != nil {
		tx.Rollback()
		return nil, nil, result.Error
	}

	tagName, err := regImg.Tag()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tag := &Tag{
		Distinction: distinction,
		ImageID:     image.ID,
		IsLatest:    true,
		IsTagged:    true,
		Name:        tagName,
	}
	result = tx.FirstOrCreate(tag, Tag{Distinction: distinction, ImageID: image.ID, Name: tagName})
	if result.Error != nil {
		tx.Rollback()
		return nil, nil, result.Error
	}

	image.Tags = append(image.Tags, tag)
	regPlatforms, err := regImg.Platforms()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	for _, p := range regPlatforms {
		regManifest, err := p.Manifest()
		if err != nil {
			tx.Rollback()
			return nil, nil, err
		}

		regManifestConfig, err := regManifest.Config()
		if err != nil {
			tx.Rollback()
			return nil, nil, err
		}

		platform := &Platform{
			Architecture:   p.Architecture(),
			CreatedAt:      time.Now().UTC(),
			Created:        time.Now().UTC(),
			ImageID:        image.ID,
			ManifestDigest: regManifestConfig.Digest().String(),
			OS:             p.OS(),
			OSVersion:      p.OSVersion(),
			Variant:        p.Variant(),
		}
		result = tx.FirstOrCreate(platform, Platform{Architecture: p.Architecture(), ImageID: image.ID, ManifestDigest: regManifestConfig.Digest().String(), OS: p.OS(), OSVersion: p.OSVersion(), Variant: p.Variant()})
		if result.Error != nil {
			tx.Rollback()
			return nil, nil, result.Error
		}

		for _, name := range p.Features() {
			feature := &Feature{
				CreatedAt: time.Now().UTC(),
				Name:      name,
			}
			result = tx.FirstOrCreate(feature, Feature{Name: name})
			if result.Error != nil {
				tx.Rollback()
				return nil, nil, result.Error
			}

			platform.Features = append(platform.Features, feature)
		}

		for _, name := range p.OSFeatures() {
			osFeature := &OSFeature{
				CreatedAt: time.Now().UTC(),
				Name:      name,
			}
			result = tx.FirstOrCreate(osFeature, OSFeature{Name: name})
			if result.Error != nil {
				tx.Rollback()
				return nil, nil, result.Error
			}

			platform.OSFeatures = append(platform.OSFeatures, osFeature)
		}

		tx.Save(platform)
		if tx.Error != nil {
			tx.Rollback()
			return nil, nil, tx.Error
		}

		for position, l := range regManifest.Layers() {
			layerDigest, err := l.Digest()
			if err != nil {
				tx.Rollback()
				return nil, nil, err
			}

			layer := &Layer{Digest: layerDigest}
			result = tx.FirstOrCreate(layer, Layer{Digest: layerDigest})
			if result.Error != nil {
				tx.Rollback()
				return nil, nil, result.Error
			}

			layerOfPlatform := &LayerOfPlatform{
				LayerID:    layer.ID,
				PlatformID: platform.ID,
				Position:   position,
			}
			result = tx.FirstOrCreate(layerOfPlatform, LayerOfPlatform{LayerID: layer.ID, PlatformID: platform.ID, Position: position})
			if result.Error != nil {
				tx.Rollback()
				return nil, nil, result.Error
			}

			platform.Layers = append(platform.Layers, layer)
		}

		image.Platforms = append(image.Platforms, platform)
	}

	tx.Commit()
	if tx.Error != nil {
		tx.Rollback()
		return nil, nil, tx.Error
	}

	return image, tag, nil
}

func (g *gormStore) Migrate() error {
	g.db.AutoMigrate(&Feature{}, &OSFeature{}, &Image{}, &Layer{}, &LayerOfPlatform{}, &Platform{}, &Tag{})
	return nil
}

func NewGormStore(connection string) (*gormStore, error) {
	db, err := gorm.Open("mysql", connection)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)
	return &gormStore{db: db}, nil
}
