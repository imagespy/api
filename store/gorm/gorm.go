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

func (g *gorm) Tags() store.TagStore {
	return &gormTag{db: g.db}
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

func (gi *gormImage) Get(o store.ImageGetOptions) (*store.Image, error) {
	if o.Digest == "" && o.ID == 0 {
		return nil, fmt.Errorf("Digest and ID not set in ImageGetOptions")
	}

	whereQuery := []string{}
	whereValues := []interface{}{}
	if o.Digest != "" {
		whereQuery = append(whereQuery, "imagespy_image.digest = ?")
		whereValues = append(whereValues, o.Digest)
	}

	if o.ID != 0 {
		whereQuery = append(whereQuery, "imagespy_image.id = ?")
		whereValues = append(whereValues, o.ID)
	}

	i := &store.Image{}
	result := gi.db.Where(strings.Join(whereQuery, " AND "), whereValues...).Take(i)
	if result.Error != nil {
		if result.Error == gormlib.ErrRecordNotFound {
			return nil, store.ErrDoesNotExist
		}

		return nil, result.Error
	}

	return i, nil
}

func (gi *gormImage) List(o store.ImageListOptions) ([]*store.Image, error) {
	imageWhereQuery := []string{}
	imageWhereValues := []interface{}{}
	if o.Digest != "" {
		imageWhereQuery = append(imageWhereQuery, "imagespy_image.digest = ?")
		imageWhereValues = append(imageWhereValues, o.Digest)
	}

	if o.Name != "" {
		imageWhereQuery = append(imageWhereQuery, "imagespy_image.name = ?")
		imageWhereValues = append(imageWhereValues, o.Name)
	}

	images := []*store.Image{}
	result := gi.db.Where(strings.Join(imageWhereQuery, " AND "), imageWhereValues...).Order("id desc").Find(&images)
	if result.Error != nil {
		return nil, result.Error
	}

	tagWhereQuery := []string{}
	tagWhereValues := []interface{}{}
	if o.TagDistinction != "" {
		tagWhereQuery = append(tagWhereQuery, "imagespy_tag.distinction = ?")
		tagWhereValues = append(tagWhereValues, o.TagDistinction)
	}

	if o.TagIsLatest != nil {
		tagWhereQuery = append(tagWhereQuery, "imagespy_tag.is_latest = ?")
		if *o.TagIsLatest {
			tagWhereValues = append(tagWhereValues, "1")
		} else {
			tagWhereValues = append(tagWhereValues, "0")
		}
	}

	if o.TagName != "" {
		tagWhereQuery = append(tagWhereQuery, "imagespy_tag.name = ?")
		tagWhereValues = append(tagWhereValues, o.TagName)
	}

	for _, image := range images {
		tmpQ := make([]string, len(tagWhereQuery))
		copy(tmpQ, tagWhereQuery)
		tmpQ = append(tmpQ, "imagespy_tag.image_id = ?")
		tmpW := make([]interface{}, len(tagWhereValues))
		copy(tmpW, tagWhereValues)
		tmpW = append(tmpW, image.ID)
		result := gi.db.Where(strings.Join(tmpQ, " AND "), tmpW...).Find(&image.Tags)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return images, nil
}

func (gi *gormImage) Update(i *store.Image) error {
	tx := gi.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := tx.Save(i)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	for _, t := range i.Tags {
		if t.ID == 0 {
			result := tx.Create(t)
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
		} else {
			result := tx.Save(t)
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

type gormTag struct {
	db *gormlib.DB
}

func (g *gormTag) Get(o store.TagGetOptions) (*store.Tag, error) {
	if o.ImageName == "" {
		return nil, fmt.Errorf("required field ImageName not set")
	}

	whereQuery := []string{}
	whereValues := []interface{}{}
	whereQuery = append(whereQuery, "imagespy_image.name = ?")
	whereValues = append(whereValues, o.ImageName)
	if o.ImageID != 0 {
		whereQuery = append(whereQuery, "imagespy_tag.image_id = ?")
		whereValues = append(whereValues, o.ImageID)
	}

	if o.Distinction != "" {
		whereQuery = append(whereQuery, "imagespy_tag.distinction = ?")
		whereValues = append(whereValues, o.Distinction)
	}

	if o.IsLatest != nil {
		whereQuery = append(whereQuery, "imagespy_tag.is_latest = ?")
		if *o.IsLatest {
			whereValues = append(whereValues, 1)
		} else {
			whereValues = append(whereValues, 0)
		}
	}

	if o.Name != "" {
		whereQuery = append(whereQuery, "imagespy_tag.name = ?")
		whereValues = append(whereValues, o.Name)
	}

	tag := &store.Tag{}
	result := g.db.Where(strings.Join(whereQuery, " AND "), whereValues...).Joins("inner join imagespy_image on imagespy_image.id = imagespy_tag.image_id").Take(tag)
	if result.Error != nil {
		if result.Error == gormlib.ErrRecordNotFound {
			return nil, store.ErrDoesNotExist
		}

		return nil, result.Error
	}

	return tag, nil
}

func (g *gormTag) Update(t *store.Tag) error {
	tx := g.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := tx.Save(t)
	if result.Error != nil {
		return result.Error
	}

	tx.Commit()
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

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
