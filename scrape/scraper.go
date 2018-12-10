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
	ScrapeImageByName(n string) error
	ScrapeLatestImage(i registry.Image) error
	ScrapeLatestImageByName(n string) error
}

func NewScraper(reg registry.Registry, s store.Store) Scraper {
	return &async{
		reg:      reg,
		store:    s,
		timeFunc: func() time.Time { return time.Now().UTC() },
	}
}

type async struct {
	reg      registry.Registry
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

func (a *async) ScrapeImageByName(n string) error {
	regImage, err := a.reg.Image(n)
	if err != nil {
		return err
	}

	err = a.ScrapeImage(regImage)
	if err != nil {
		return err
	}

	return nil
}

func (a *async) ScrapeLatestImage(i registry.Image) error {
	start := time.Now()
	defer func() { promScrapeLatestDuration.Observe(time.Since(start).Seconds()) }()
	regImgTag, err := i.Tag()
	if err != nil {
		return err
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
		return err
	}
	var currentImage *store.Image
	if currentTag != nil {
		currentImage, err = a.store.Images().Get(store.ImageGetOptions{ID: currentTag.ImageID})
		if err != nil {
			return err
		}
	}

	regImages, err := i.Repository().Images()
	if err != nil {
		return err
	}

	latestRegImage := i
	for _, regImageItem := range regImages {
		currentImageTag, err := regImageItem.Tag()
		if err != nil {
			return err
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
		return err
	}

	var latestImage *store.Image
	latestImage, err = a.store.Images().Get(store.ImageGetOptions{Digest: latestRegImageDigest})
	if err != nil {
		if err == store.ErrDoesNotExist {
			var latestImageLayers []*store.Layer
			latestImage, latestImageLayers, err = a.CreateStoreImageFromRegistryImage(latestVP.Distinction(), latestRegImage)
			if err != nil {
				return err
			}

			for _, l := range latestImageLayers {
				err := a.updateSourceImagesOfLayer(l)
				if err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}

	if currentImage != nil && currentImage.Digest == latestImage.Digest {
		return nil
	}

	latestTag, err := a.store.Tags().Get(store.TagGetOptions{
		Distinction: latestVP.Distinction(),
		ImageID:     latestImage.ID,
		ImageName:   latestImage.Name,
		Name:        latestVP.String(),
	})
	if err != nil {
		return err
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
		err := a.store.Tags().Update(currentTag)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *async) ScrapeLatestImageByName(n string) error {
	regImage, err := a.reg.Image(n)
	if err != nil {
		return err
	}

	err = a.ScrapeLatestImage(regImage)
	if err != nil {
		return err
	}

	return nil
}

func (a *async) CreateStoreImageFromRegistryImage(distinction string, regImg registry.Image) (*store.Image, []*store.Layer, error) {
	imageDigest, err := regImg.Digest()
	if err != nil {
		return nil, nil, err
	}

	imageSV, err := regImg.SchemaVersion()
	if err != nil {
		return nil, nil, err
	}

	image := &store.Image{
		CreatedAt:     a.timeFunc(),
		Digest:        imageDigest,
		Name:          regImg.Repository().FullName(),
		SchemaVersion: imageSV,
		ScrapedAt:     a.timeFunc(),
	}
	err = a.store.Images().Create(image)
	if err != nil {
		return nil, nil, err
	}

	tagName, err := regImg.Tag()
	if err != nil {
		return nil, nil, err
	}

	tag := &store.Tag{
		Distinction: distinction,
		ImageID:     image.ID,
		IsLatest:    false,
		IsTagged:    true,
		Name:        tagName,
	}
	err = a.store.Tags().Create(tag)
	if err != nil {
		return nil, nil, err
	}

	regPlatforms, err := regImg.Platforms()
	if err != nil {
		return nil, nil, err
	}

	layers := []*store.Layer{}
	layerClient := a.store.Layers()
	layerPositionClient := a.store.LayerPositions()
	platformClient := a.store.Platforms()
	for _, p := range regPlatforms {
		regManifest, err := p.Manifest()
		if err != nil {
			return nil, nil, err
		}

		regManifestConfig, err := regManifest.Config()
		if err != nil {
			return nil, nil, err
		}

		platform := &store.Platform{
			Architecture:   p.Architecture(),
			CreatedAt:      a.timeFunc(),
			Created:        a.timeFunc(),
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
			continue
		}

		for idx, l := range regManifest.Layers() {
			layerDigest, err := l.Digest()
			if err != nil {
				return nil, nil, err
			}

			layer := &store.Layer{Digest: layerDigest}
			err = layerClient.Create(layer)
			if err != nil {
				log.Errorf("unable to create layer %s for platform %s: %s", layer.Digest, platform.ManifestDigest, err)
				break
			}

			layerPosition := &store.LayerPosition{LayerID: layer.ID, PlatformID: platform.ID, Position: idx}
			err = layerPositionClient.Create(layerPosition)
			if err != nil {
				log.Errorf("unable to create layer position '%d' for layer '%s' for platform '%s': %s", idx, layer.Digest, platform.ManifestDigest, err)
				break
			}

			layers = append(layers, layer)
		}
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
