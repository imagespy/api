package scrape

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/versionparser"
	log "github.com/sirupsen/logrus"
)

var (
	promScrapeDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:      "scrape_duration_seconds",
			Namespace: "imagespy",
			Help:      "A histogram of the time it took to scrape an image.",
			Buckets:   []float64{.5, 1, 5, 10},
		},
	)

	promScrapeLatestDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:      "scrape_latest_duration_seconds",
			Namespace: "imagespy",
			Help:      "A histogram of the time it took to scrape the latest version of an image.",
			Buckets:   []float64{.5, 1, 5, 10},
		},
	)
)

type Scraper interface {
	ScrapeImage(i registry.Image) error
	ScrapeLatestImage(i registry.Image) error
}

func NewScraper(s store.Store) Scraper {
	return &async{
		store:    s,
		timeFunc: func() time.Time { return time.Now().UTC() },
	}
}

type async struct {
	store    store.Store
	timeFunc func() time.Time
}

func (a *async) ScrapeImage(i registry.Image) error {
	start := time.Now()
	defer func() { promScrapeDuration.Observe(time.Since(start).Seconds()) }()
	digest, err := i.Digest()
	if err != nil {
		return fmt.Errorf("ScrapeImage: retrieving digest failed: %s", err)
	}

	tagRef, err := i.Tag()
	if err != nil {
		return fmt.Errorf("ScrapeImage: retrieving tag failed: %s", err)
	}

	vp := versionparser.FindForVersion(tagRef)
	image, err := a.store.Images().Get(store.ImageGetOptions{Digest: digest})
	if err == nil {
		newTag := &store.Tag{
			Distinction: vp.Distinction(),
			ImageID:     image.ID,
			IsLatest:    false,
			IsTagged:    true,
			Name:        tagRef,
		}

		tags, err := a.store.Tags().List(store.TagListOptions{ImageID: image.ID})
		if err != nil {
			return err
		}

		tagExists := false
		for _, tagItem := range tags {
			if tagItem.Name == newTag.Name {
				tagExists = true
			}
		}

		if !tagExists {
			err := a.store.Tags().Create(newTag)
			if err != nil {
				return err
			}
		}

		image.ScrapedAt = a.timeFunc()
		err = a.store.Images().Update(image)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil && err == store.ErrDoesNotExist {
		_, layers, err := a.CreateStoreImageFromRegistryImage(vp.Distinction(), i)
		if err != nil {
			return err
		}

		for _, l := range layers {
			err := a.updateSourceImagesOfLayer(l)
			if err != nil {
				log.Errorf("failed to update source image of layer %d: %s", l.ID, err)
			}
		}

		return nil
	}

	return fmt.Errorf("ScrapeImage: reading image with digest %s failed: %s", digest, err)
}

func (a *async) ScrapeLatestImage(i registry.Image) error {
	start := time.Now()
	defer func() { promScrapeLatestDuration.Observe(time.Since(start).Seconds()) }()
	regImgTag, err := i.Tag()
	if err != nil {
		return fmt.Errorf("ScrapeLatestImage - getting tag of registry image: %s", err)
	}

	latestVP := versionparser.FindForVersion(regImgTag)

	b := true
	currentTag, err := a.store.Tags().Get(store.TagGetOptions{
		Distinction: latestVP.Distinction(),
		ImageName:   i.Repository().FullName(),
		IsLatest:    &b,
		Name:        regImgTag,
	})
	if err != nil && err != store.ErrDoesNotExist {
		return fmt.Errorf("ScrapeLatestImage - getting tag of image: %s", err)
	}
	var currentImage *store.Image
	if currentTag != nil {
		currentImage, err = a.store.Images().Get(store.ImageGetOptions{ID: currentTag.ImageID})
		if err != nil {
			return fmt.Errorf("ScrapeLatestImage - getting image with id %d: %s", currentTag.ImageID, err)
		}
	}

	regImages, err := i.Repository().Images()
	if err != nil {
		return fmt.Errorf("ScrapeLatestImage - getting images of registry repository: %s", err)
	}

	latestRegImage := i
	for _, regImageItem := range regImages {
		currentImageTag, err := regImageItem.Tag()
		if err != nil {
			return fmt.Errorf("ScrapeLatestImage - getting tag of registry image %s: %s", regImageItem.Repository().FullName(), err)
		}

		currentVP := versionparser.FindForVersion(currentImageTag)
		if currentVP.Distinction() != latestVP.Distinction() {
			continue
		}

		currentIsGreater, err := currentVP.IsGreaterThan(latestVP)
		if err != nil {
			continue
		}

		if currentIsGreater {
			latestVP = currentVP
			latestRegImage = regImageItem
		}
	}

	latestRegImageDigest, err := latestRegImage.Digest()
	if err != nil {
		return fmt.Errorf("ScrapeLatestImage - getting digest of latest registry image %s: %s", latestRegImage.Repository().FullName(), err)
	}

	latestImageCreated := false
	var latestImage *store.Image
	latestImage, err = a.store.Images().Get(store.ImageGetOptions{Digest: latestRegImageDigest})
	if err != nil {
		if err == store.ErrDoesNotExist {
			var latestImageLayers []*store.Layer
			latestImage, latestImageLayers, err = a.CreateStoreImageFromRegistryImage(latestVP.Distinction(), latestRegImage)
			if err != nil {
				return fmt.Errorf("ScrapeLatestImage - creating image from registry image %s, distinction %s: %s", latestRegImage.Repository().FullName(), latestVP.Distinction(), err)
			}

			for _, l := range latestImageLayers {
				err := a.updateSourceImagesOfLayer(l)
				if err != nil {
					return fmt.Errorf("ScrapeLatestImage - updating source images of layer %s: %s", l.Digest, err)
				}
			}
			latestImageCreated = true
		} else {
			return fmt.Errorf("ScrapeLatestImage - getting latest image by digest %s: %s", latestRegImageDigest, err)
		}
	}

	if currentImage != nil && currentImage.Digest == latestImage.Digest {
		currentImage.ScrapedAt = a.timeFunc()
		err = a.store.Images().Update(currentImage)
		if err != nil {
			return err
		}

		return nil
	}

	latestTag, err := a.store.Tags().Get(store.TagGetOptions{
		Distinction: latestVP.Distinction(),
		ImageID:     latestImage.ID,
		ImageName:   latestImage.Name,
		Name:        latestVP.String(),
	})
	if err != nil {
		if err == store.ErrDoesNotExist {
			latestTag = &store.Tag{
				Distinction: latestVP.Distinction(),
				ImageID:     latestImage.ID,
				IsLatest:    true,
				IsTagged:    true,
				Name:        latestVP.String(),
			}

			err := a.store.Tags().Create(latestTag)
			if err != nil {
				return fmt.Errorf("ScrapeLatestImage - creating latest tag %s for image %s: %s", latestTag.Name, latestImage.Name, err)
			}
		} else {
			return fmt.Errorf("ScrapeLatestImage - getting latest tag with distinction %s - %s: %s", latestVP.Distinction(), latestVP.String(), err)
		}
	}

	if latestTag.IsLatest == false {
		latestTag.IsLatest = true
		err := a.store.Tags().Update(latestTag)
		if err != nil {
			return err
		}
	}

	if currentTag != nil && currentTag.IsLatest == true {
		currentTag.IsLatest = false
		if currentTag.Name == latestTag.Name {
			currentTag.IsTagged = false
		}

		err := a.store.Tags().Update(currentTag)
		if err != nil {
			return err
		}
	}

	if !latestImageCreated {
		latestImage.ScrapedAt = a.timeFunc()
		err = a.store.Images().Update(latestImage)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *async) CreateStoreImageFromRegistryImage(distinction string, regImg registry.Image) (*store.Image, []*store.Layer, error) {
	tx, err := a.store.Transaction()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	imageDigest, err := regImg.Digest()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	imageSV, err := regImg.SchemaVersion()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	image := &store.Image{
		CreatedAt:     a.timeFunc(),
		Digest:        imageDigest,
		Name:          regImg.Repository().FullName(),
		SchemaVersion: imageSV,
		ScrapedAt:     a.timeFunc(),
	}
	err = tx.Images().Create(image)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tagName, err := regImg.Tag()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tag := &store.Tag{
		Distinction: distinction,
		ImageID:     image.ID,
		IsLatest:    false,
		IsTagged:    true,
		Name:        tagName,
	}
	err = tx.Tags().Create(tag)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	regPlatforms, err := regImg.Platforms()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	layers := []*store.Layer{}
	layerClient := tx.Layers()
	layerPositionClient := tx.LayerPositions()
	platformClient := tx.Platforms()
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

		history, err := regManifestConfig.History()
		if err != nil {
			tx.Rollback()
			return nil, nil, err
		}

		platform := &store.Platform{
			Architecture:   p.Architecture(),
			CreatedAt:      a.timeFunc(),
			Created:        history[len(history)-1].Created,
			ImageID:        image.ID,
			ManifestDigest: regManifestConfig.Digest().String(),
			OS:             p.OS(),
			OSVersion:      p.OSVersion(),
			Variant:        p.Variant(),
		}

		for _, name := range p.Features() {
			platform.Features = append(platform.Features, &store.Feature{
				CreatedAt: a.timeFunc(),
				Name:      name,
			})
		}

		for _, name := range p.OSFeatures() {
			platform.OSFeatures = append(platform.OSFeatures, &store.OSFeature{
				CreatedAt: a.timeFunc(),
				Name:      name,
			})
		}

		err = platformClient.Create(platform)
		if err != nil {
			log.Errorf("unable to create platform %s for image %d: %s", platform.ManifestDigest, image.ID, err)
			tx.Rollback()
			return nil, nil, err
		}

		for idx, l := range regManifest.Layers() {
			layerDigest, err := l.Digest()
			if err != nil {
				tx.Rollback()
				return nil, nil, err
			}

			layer := &store.Layer{Digest: layerDigest}
			err = layerClient.Create(layer)
			if err != nil {
				log.Errorf("unable to create layer %s for platform %s: %s", layer.Digest, platform.ManifestDigest, err)
				tx.Rollback()
				return nil, nil, err
			}

			layerPosition := &store.LayerPosition{LayerID: layer.ID, PlatformID: platform.ID, Position: idx}
			err = layerPositionClient.Create(layerPosition)
			if err != nil {
				log.Errorf("unable to create layer position '%d' for layer '%s' for platform '%s': %s", idx, layer.Digest, platform.ManifestDigest, err)
				tx.Rollback()
				return nil, nil, err
			}

			layers = append(layers, layer)
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
	}

	return image, layers, nil
}

func (a *async) updateSourceImagesOfLayer(l *store.Layer) error {
	platforms, err := a.store.Platforms().List(store.PlatformListOptions{LayerDigest: l.Digest})
	if err != nil {
		return err
	}

	layerClient := a.store.Layers()
	layerPositionClient := a.store.LayerPositions()
	length := 1000
	var sourcePlatforms []*store.Platform
	for _, p := range platforms {
		layerPositions, err := layerPositionClient.List(store.LayerPositionListOptions{PlatformID: p.ID})
		if err != nil {
			return err
		}

		if len(layerPositions) < length {
			length = len(layerPositions)
			sourcePlatforms = []*store.Platform{p}
		} else if len(layerPositions) == length {
			sourcePlatforms = append(sourcePlatforms, p)
		}
	}

	sourceImageIDs, needsUpdate := newSourceImageIDs(l, sourcePlatforms)
	if needsUpdate {
		l.SourceImageIDs = sourceImageIDs
		err = layerClient.Update(l)
		if err != nil {
			return err
		}
	}

	return nil
}

func newSourceImageIDs(l *store.Layer, platforms []*store.Platform) ([]int, bool) {
	sourceImageIDsCurrent := map[int]struct{}{}
	for _, siid := range l.SourceImageIDs {
		sourceImageIDsCurrent[siid] = struct{}{}
	}

	needsUpdate := false
	sourceImageIDsNew := map[int]struct{}{}
	for _, p := range platforms {
		_, ok := sourceImageIDsCurrent[p.ImageID]
		if !ok {
			needsUpdate = true
		}

		sourceImageIDsNew[p.ImageID] = struct{}{}
	}

	if !needsUpdate {
		return l.SourceImageIDs, false
	}

	sourceImageIDs := []int{}
	for imageID := range sourceImageIDsNew {
		sourceImageIDs = append(sourceImageIDs, imageID)
	}

	return sourceImageIDs, true
}
