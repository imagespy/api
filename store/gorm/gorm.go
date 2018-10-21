package gorm

import (
	"fmt"
	"strings"

	"github.com/imagespy/api/store"
	gormlib "github.com/jinzhu/gorm"
)

type gorm struct {
	db *gormlib.DB
}

func (g *gorm) Images() store.ImageStore {
	return &gormImage{db: g.db}
}

func (g *gorm) Close() error {
	return g.db.Close()
}

type gormImage struct {
	db *gormlib.DB
}

func (gi *gormImage) Create(i *store.Image) error {
	if i.ID != 0 {
		return fmt.Errorf("Image already created")
	}

	tx := gi.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := tx.FirstOrCreate(i, store.Image{Digest: i.Digest, Name: i.Name, SchemaVersion: i.SchemaVersion})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	for _, t := range i.Tags {
		result = tx.FirstOrCreate(t, store.Tag{Distinction: t.Distinction, ImageID: i.ID, Name: t.Name})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}

	for _, p := range i.Platforms {
		result = tx.FirstOrCreate(p, store.Platform{Architecture: p.Architecture, ImageID: i.ID, ManifestDigest: p.ManifestDigest, OS: p.OS, OSVersion: p.OSVersion, Variant: p.Variant})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		for _, pf := range p.Features {
			result = tx.FirstOrCreate(pf, store.Feature{Name: pf.Name})
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
		}

		for _, posf := range p.OSFeatures {
			result = tx.FirstOrCreate(posf, store.OSFeature{Name: posf.Name})
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
		}

		result = tx.Save(p)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		for position, l := range p.Layers {
			result = tx.FirstOrCreate(l, store.Layer{Digest: l.Digest})
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
			fmt.Printf("l: %d - p: %d\n", l.ID, p.ID)

			layerOfPlatform := &store.LayerOfPlatform{
				LayerID:    l.ID,
				PlatformID: p.ID,
				Position:   position,
			}
			result = tx.FirstOrCreate(layerOfPlatform, store.LayerOfPlatform{LayerID: l.ID, PlatformID: p.ID, Position: position})
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
		}
	}

	tx.Commit()
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	return nil
}

func (gi *gormImage) Get(digest string) (*store.Image, error) {
	i := &store.Image{}
	result := gi.db.Where("imagespy_image.digest = ?", digest).Preload("Tags").Preload("Platforms")
	if result.Error != nil {
		if result.Error == gormlib.ErrRecordNotFound {
			return nil, store.ErrDoesNotExist
		}

		return nil, result.Error
	}

	return i, nil
}

func (gi *gormImage) List(o store.ImageListOptions) ([]*store.Image, error) {
	whereQuery := []string{}
	whereValues := []interface{}{}
	if o.Digest != "" {
		whereQuery = append(whereQuery, "imagespy_image.digest = ?")
		whereValues = append(whereValues, o.Digest)
	}

	if o.Name != "" {
		whereQuery = append(whereQuery, "imagespy_image.name = ?")
		whereValues = append(whereValues, o.Name)
	}

	if o.TagDistinction != "" {
		whereQuery = append(whereQuery, "imagespy_tag.distinction = ?")
		whereValues = append(whereValues, o.TagDistinction)
	}

	if o.TagIsLatest != nil {
		whereQuery = append(whereQuery, "imagespy_tag.is_latest = ?")
		if *o.TagIsLatest {
			whereValues = append(whereValues, "1")
		} else {
			whereValues = append(whereValues, "0")
		}
	}

	if o.TagName != "" {
		whereQuery = append(whereQuery, "imagespy_tag.name = ?")
		whereValues = append(whereValues, o.TagName)
	}

	i := []*store.Image{}
	result := gi.db.Where(strings.Join(whereQuery, " AND "), whereValues...).Joins("inner join imagespy_tag on imagespy_tag.image_id = imagespy_image.id").Order("id desc").Preload("Tags").Find(&i)
	if result.Error != nil {
		if result.Error == gormlib.ErrRecordNotFound {
			return nil, store.ErrDoesNotExist
		}

		return nil, result.Error
	}

	return i, nil
}

func (gi *gormImage) Update(i *store.Image) error {
	tx := gi.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := gi.db.Save(i)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	for _, t := range i.Tags {
		if t.ID == 0 {
			result := gi.db.Create(t)
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
		} else {
			result := gi.db.Save(t)
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
		}
	}

	tx.Commit()
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	return nil
}

func (g *gorm) Migrate() error {
	g.db.AutoMigrate(&store.Feature{}, &store.OSFeature{}, &store.Image{}, &store.Layer{}, &store.LayerOfPlatform{}, &store.Platform{}, &store.Tag{})
	return nil
}

func New(connection string) (*gorm, error) {
	db, err := gormlib.Open("mysql", connection)
	if err != nil {
		return nil, err
	}

	// db.LogMode(true)
	return &gorm{db: db}, nil
}
