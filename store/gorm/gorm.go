package gorm

import (
	"fmt"
	"strings"

	"github.com/imagespy/api/store"
	gormlib "github.com/jinzhu/gorm"
)

type sourceImageOfLayer struct {
	store.Model
	ImageID int
	LayerID int
}

func (sourceImageOfLayer) TableName() string {
	return "imagespy_layer_source_images"
}

type gorm struct {
	db *gormlib.DB
}

func (g *gorm) Images() store.ImageStore {
	return &gormImage{db: g.db}
}

func (g *gorm) Layers() store.LayerStore {
	return &gormLayer{db: g.db}
}

func (g *gorm) LayerPositions() store.LayerPositionStore {
	return &gormLayerPosition{db: g.db}
}

func (g *gorm) Platforms() store.PlatformStore {
	return &gormPlatform{db: g.db}
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

	tx.Commit()
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	return nil
}

func (gi *gormImage) Get(o store.ImageGetOptions) (*store.Image, error) {
	if o.Digest == "" && o.ID == 0 && o.Name == "" {
		return nil, fmt.Errorf("Digest, ID and Name not set in ImageGetOptions")
	}

	joinWithTag := false
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

	if o.Name != "" {
		whereQuery = append(whereQuery, "imagespy_image.name = ?")
		whereValues = append(whereValues, o.Name)
	}

	if o.TagDistinction != "" {
		whereQuery = append(whereQuery, "imagespy_tag.distinction = ?")
		whereValues = append(whereValues, o.TagDistinction)
		joinWithTag = true
	}

	if o.TagIsLatest != nil {
		whereQuery = append(whereQuery, "imagespy_tag.is_latest = ?")
		if *o.TagIsLatest {
			whereValues = append(whereValues, "1")
		} else {
			whereValues = append(whereValues, "0")
		}
		joinWithTag = true
	}

	if o.TagName != "" {
		whereQuery = append(whereQuery, "imagespy_tag.name = ?")
		whereValues = append(whereValues, o.TagName)
		joinWithTag = true
	}

	q := gi.db
	if joinWithTag {
		q = q.Joins("inner join imagespy_tag on imagespy_tag.image_id = imagespy_image.id")
	}

	i := &store.Image{}
	result := q.Where(strings.Join(whereQuery, " AND "), whereValues...).Order("created_at desc").Take(i)
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

	tx.Commit()
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	return nil
}

type gormLayer struct {
	db *gormlib.DB
}

func (g *gormLayer) Create(l *store.Layer) error {
	tx := g.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := tx.FirstOrCreate(l, store.Layer{Digest: l.Digest})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	for _, imageID := range l.SourceImageIDs {
		siid := &sourceImageOfLayer{
			ImageID: imageID,
			LayerID: l.ID,
		}

		result := tx.FirstOrCreate(siid, sourceImageOfLayer{ImageID: imageID, LayerID: l.ID})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}

	result = tx.Commit()
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	return nil
}

func (g *gormLayer) Get(o store.LayerGetOptions) (*store.Layer, error) {
	whereQuery := []string{}
	whereValues := []interface{}{}
	if o.Digest != "" {
		whereQuery = append(whereQuery, "imagespy_layer.digest = ?")
		whereValues = append(whereValues, o.Digest)
	}

	if o.ID != 0 {
		whereQuery = append(whereQuery, "imagespy_layer.id = ?")
		whereValues = append(whereValues, o.ID)
	}

	layer := &store.Layer{}
	result := g.db.Where(strings.Join(whereQuery, " AND "), whereValues...).First(layer)
	if result.Error != nil {
		if result.Error == gormlib.ErrRecordNotFound {
			return nil, store.ErrDoesNotExist
		}

		return nil, result.Error
	}

	sourceImageIDs := []*sourceImageOfLayer{}
	siolResult := g.db.Where("imagespy_layer_source_images.layer_id = ?", layer.ID).Find(&sourceImageIDs)
	if siolResult.Error != nil {
		return nil, siolResult.Error
	}

	for _, sid := range sourceImageIDs {
		layer.SourceImageIDs = append(layer.SourceImageIDs, sid.ImageID)
	}

	return layer, nil
}

func (g *gormLayer) List(o store.LayerListOptions) ([]*store.Layer, error) {
	query := g.db
	if o.PlatformID != 0 {
		query = query.Joins("inner join imagespy_layerofplatform on imagespy_layerofplatform.layer_id = imagespy_layer.id").
			Where("imagespy_layerofplatform.platform_id = ?", o.PlatformID)
	}

	layers := []*store.Layer{}
	result := query.Find(&layers)
	if result.Error != nil {
		return nil, result.Error
	}

	type sourceImageIDsResult struct {
		ImageID int
	}

	for _, l := range layers {
		rows, err := g.db.Raw("select image_id from imagespy_layer_source_images where layer_id = ?", l.ID).Rows()
		if err != nil {
			rows.Close()
			return nil, err
		}

		for rows.Next() {
			var result sourceImageIDsResult
			err := g.db.ScanRows(rows, &result)
			if err != nil {
				rows.Close()
				return nil, nil
			}

			l.SourceImageIDs = append(l.SourceImageIDs, result.ImageID)
		}

		rows.Close()
	}

	return layers, nil
}

func (g *gormLayer) Update(l *store.Layer) error {
	tx := g.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := tx.Exec("delete from imagespy_layer_source_images where layer_id = ?", l.ID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	for _, imageID := range l.SourceImageIDs {
		result := tx.Exec("insert into imagespy_layer_source_images (image_id, layer_id) VALUES (?, ?)", imageID, l.ID)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}

	result = tx.Commit()
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	return nil
}

type gormLayerPosition struct {
	db *gormlib.DB
}

func (g *gormLayerPosition) Create(lp *store.LayerPosition) error {
	result := g.db.FirstOrCreate(lp, store.LayerPosition{LayerID: lp.LayerID, PlatformID: lp.PlatformID, Position: lp.Position})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (g *gormLayerPosition) List(o store.LayerPositionListOptions) ([]*store.LayerPosition, error) {
	whereQuery := []string{}
	whereValues := []interface{}{}
	if o.LayerID != 0 {
		whereQuery = append(whereQuery, "imagespy_layerofplatform.layer_id = ?")
		whereValues = append(whereValues, o.LayerID)
	}

	if o.PlatformID != 0 {
		whereQuery = append(whereQuery, "imagespy_layerofplatform.platform_id = ?")
		whereValues = append(whereValues, o.PlatformID)
	}

	lp := []*store.LayerPosition{}
	result := g.db.Where(strings.Join(whereQuery, " AND "), whereValues...).Order("imagespy_layerofplatform.position asc").Find(&lp)
	if result.Error != nil {
		return nil, result.Error
	}

	return lp, nil
}

type gormPlatform struct {
	db *gormlib.DB
}

func (g *gormPlatform) Create(p *store.Platform) error {
	result := g.db.Create(p)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (g *gormPlatform) Get(o store.PlatformGetOptions) (*store.Platform, error) {
	whereQuery := []string{}
	whereValues := []interface{}{}
	if o.Architecture != "" {
		whereQuery = append(whereQuery, "imagespy_platform.architecture = ?")
		whereValues = append(whereValues, o.Architecture)
	}

	if o.ImageID != 0 {
		whereQuery = append(whereQuery, "imagespy_platform.image_id = ?")
		whereValues = append(whereValues, o.ImageID)
	}

	if o.OS != "" {
		whereQuery = append(whereQuery, "imagespy_platform.os = ?")
		whereValues = append(whereValues, o.OS)
	}

	whereQuery = append(whereQuery, "imagespy_platform.os_version = ?")
	whereValues = append(whereValues, o.OSVersion)
	whereQuery = append(whereQuery, "imagespy_platform.variant = ?")
	whereValues = append(whereValues, o.Variant)

	p := &store.Platform{}
	result := g.db.Where(strings.Join(whereQuery, " AND "), whereValues...).Take(p)
	if result.Error != nil {
		if result.Error == gormlib.ErrRecordNotFound {
			return nil, store.ErrDoesNotExist
		}

		return nil, result.Error
	}

	return p, nil
}

func (g *gormPlatform) List(o store.PlatformListOptions) ([]*store.Platform, error) {
	platforms := []*store.Platform{}
	query := g.db
	if o.LayerDigest != "" {
		query = query.Joins("inner join imagespy_layerofplatform on imagespy_layerofplatform.platform_id = imagespy_platform.id").
			Joins("inner join imagespy_layer on imagespy_layer.id = imagespy_layerofplatform.layer_id").
			Where("imagespy_layer.digest = ?", o.LayerDigest)
	}

	result := query.Find(&platforms)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, p := range platforms {
		features := []*store.Feature{}
		featuresResult := g.db.Joins("inner join imagespy_platform_features on imagespy_platform_features.feature_id = imagespy_feature.id").
			Where("imagespy_platform_features.platform_id = ?", p.ID).
			Find(&features)
		if featuresResult.Error != nil {
			return nil, featuresResult.Error
		}

		p.Features = features
		osFeatures := []*store.OSFeature{}
		osFeaturesResult := g.db.Joins("inner join imagespy_platform_os_features on imagespy_platform_os_features.osfeature_id = imagespy_osfeature.id").
			Where("imagespy_platform_os_features.platform_id = ?", p.ID).
			Find(&osFeatures)
		if osFeaturesResult.Error != nil {
			return nil, osFeaturesResult.Error
		}

		p.OSFeatures = osFeatures
	}

	return platforms, nil
}

type gormTag struct {
	db *gormlib.DB
}

func (g *gormTag) Create(tag *store.Tag) error {
	result := g.db.Create(tag)
	if result.Error != nil {
		return result.Error
	}

	return nil
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

func (g *gormTag) List(o store.TagListOptions) ([]*store.Tag, error) {
	tags := []*store.Tag{}
	whereQuery := []string{}
	whereValues := []interface{}{}
	if o.ImageID != 0 {
		whereQuery = append(whereQuery, "imagespy_tag.image_id = ?")
		whereValues = append(whereValues, o.ImageID)
	}

	if o.IsLatest != nil {
		whereQuery = append(whereQuery, "imagespy_tag.is_latest = ?")
		if *o.IsLatest {
			whereValues = append(whereValues, 1)
		} else {
			whereValues = append(whereValues, 0)
		}
	}

	result := g.db.Where(strings.Join(whereQuery, " AND "), whereValues...).Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}

	return tags, nil
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
