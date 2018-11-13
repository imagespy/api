package scrape

import (
	"fmt"
	"time"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/versionparser"
	log "github.com/sirupsen/logrus"
)

type Work struct {
	Address    string
	Repository string
	Tags       []string
}

type Scraper interface {
	ScrapeImage(i registry.Image) error
	ScrapeLatestImage(i registry.Image) error
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

func (a *async) ScrapeLatestImageForImages(imageNames []string) {
	for _, name := range imageNames {
		regImg, err := a.reg.Image(name)
		if err != nil {
			log.Errorf("retrieve image %s: %s", name, err)
			continue
		}

		err = a.ScrapeLatestImage(regImg)
		if err != nil {
			log.Error(err)
			continue
		}
	}
}

func (a *async) ScrapeImage(i registry.Image) error {
	digest, err := i.Digest()
	if err != nil {
		return fmt.Errorf("ScrapeImage: retrieving digest failed: %s", err)
	}

	tagRef, err := i.Tag()
	if err != nil {
		return fmt.Errorf("ScrapeImage: retrieving tag failed: %s", err)
	}

	vp := versionparser.FindForVersion(tagRef)
	image, err := a.store.Images().Get(store.ImageGetOptions{
		Digest: digest,
	})
	if err == nil {
		newTag := &store.Tag{
			Distinction: vp.Distinction(),
			IsLatest:    false,
			IsTagged:    true,
			Name:        tagRef,
		}
		if image.HasTag(newTag) == false {
			image.Tags = append(image.Tags, newTag)
			err := a.store.Images().Update(image)
			if err != nil {
				return fmt.Errorf("ScrapeImage: updating image failed: %s", err)
			}
		}

		return nil
	}

	if err != nil && err == store.ErrDoesNotExist {
		image, err := a.CreateStoreImageFromRegistryImage(vp.Distinction(), i)
		if err != nil {
			return err
		}

		err = a.store.Images().Create(image)
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("ScrapeImage: reading image with digest %s failed: %s", digest, err)
}

func (a *async) ScrapeLatestImage(i registry.Image) error {
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
		currentImage, err = a.store.Images().Get(store.ImageGetOptions{
			ID: currentTag.ImageID,
		})
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
	latestImage, err = a.store.Images().Get(store.ImageGetOptions{
		Digest: latestRegImageDigest,
	})
	if err != nil {
		if err == store.ErrDoesNotExist {
			latestImage, err = a.CreateStoreImageFromRegistryImage(latestVP.Distinction(), latestRegImage)
			if err != nil {
				return err
			}

			err = a.store.Images().Create(latestImage)
			if err != nil {
				return err
			}

			for _, p := range latestImage.Platforms {
				for _, l := range p.Layers {
					err := a.updateSourceImagesOfLayer(l)
					if err != nil {
						return err
					}
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

// func (s *async) scrapeLatestImages() error {
// 	tags, err := s.store.FindAllLatestTagsWithImage()
// 	if err != nil {
// 		return err
// 	}
//
// 	imagesByName := map[string][]*store.Tag{}
// 	for _, tag := range tags {
// 		_, ok := imagesByName[tag.Image.Name]
// 		if ok {
// 			imagesByName[tag.Image.Name] = append(imagesByName[tag.Image.Name], tag)
// 		} else {
// 			imagesByName[tag.Image.Name] = []*store.Tag{tag}
// 		}
// 	}
//
// 	for imageName, tagsToScrape := range imagesByName {
// 		regRepo, err := registry.NewRepository(imageName, false)
// 		if err != nil {
// 			return err
// 		}
//
// 		for _, t := range tagsToScrape {
// 			regImage := regRepo.Image("", t.Name)
// 			latestImage, err := ScrapeLatestImageOfRegistryImage(regImage, s.store)
// 			if err != nil {
// 				return err
// 			}
//
// 			if latestImage.Digest != t.Image.Digest {
// 				t.IsLatest = false
// 				err = s.store.UpdateTag(t)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}
//
// 	return nil
// }

func (a *async) CreateStoreImageFromRegistryImage(distinction string, regImg registry.Image) (*store.Image, error) {
	imageDigest, err := regImg.Digest()
	if err != nil {
		return nil, err
	}

	imageSV, err := regImg.SchemaVersion()
	if err != nil {
		return nil, err
	}

	image := &store.Image{
		CreatedAt:     a.timeFunc(),
		Digest:        imageDigest,
		Name:          regImg.Repository().FullName(),
		SchemaVersion: imageSV,
		ScrapedAt:     a.timeFunc(),
	}
	tagName, err := regImg.Tag()
	if err != nil {
		return nil, err
	}

	tag := &store.Tag{
		Distinction: distinction,
		IsLatest:    false,
		IsTagged:    true,
		Name:        tagName,
	}
	if image.HasTag(tag) == false {
		image.Tags = append(image.Tags, tag)
	}

	regPlatforms, err := regImg.Platforms()
	if err != nil {
		return nil, err
	}

	for _, p := range regPlatforms {
		regManifest, err := p.Manifest()
		if err != nil {
			return nil, err
		}

		regManifestConfig, err := regManifest.Config()
		if err != nil {
			return nil, err
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

		for _, l := range regManifest.Layers() {
			layerDigest, err := l.Digest()
			if err != nil {
				return nil, err
			}

			platform.Layers = append(platform.Layers, &store.Layer{Digest: layerDigest})
		}

		image.Platforms = append(image.Platforms, platform)
	}

	return image, nil
}

func (a *async) updateSourceImagesOfLayer(l *store.Layer) error {
	layerClient := a.store.Layers()
	platforms, err := a.store.Platforms().List(store.PlatformListOptions{LayerDigest: l.Digest})
	if err != nil {
		return err
	}

	length := 1000
	var sourcePlatforms []*store.Platform
	for _, p := range platforms {
		layers, err := layerClient.List(store.LayerListOptions{PlatformID: p.ID})
		if err != nil {
			return err
		}

		if len(layers) < length {
			length = len(layers)
			sourcePlatforms = []*store.Platform{p}
		} else if len(layers) == length {
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
