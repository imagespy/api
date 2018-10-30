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
	ScrapeLatestImageForImages(imageNames []string)
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
	image, err := a.store.Images().Get(digest)
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

	b := true
	images, err := a.store.Images().List(store.ImageListOptions{
		Name:        i.Repository().FullName(),
		TagIsLatest: &b,
		TagName:     regImgTag,
	})
	if err != nil {
		return err
	}

	if len(images) > 1 {
		return fmt.Errorf("query for latest tag of %s:%s returned ambiguous result: expected 1 got %d", i.Repository().FullName(), regImgTag, len(images))
	}

	var currentImage *store.Image
	var currentTag *store.Tag
	if len(images) != 0 {
		currentImage = images[0]
		currentTag, err = currentImage.FindTag(regImgTag)
		if err != nil {
			return err
		}
	}

	regImages, err := i.Repository().Images()
	if err != nil {
		return err
	}

	latestVP := versionparser.FindForVersion(regImgTag)
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
	if currentImage == nil {
		latestImage, err = a.CreateStoreImageFromRegistryImage(latestVP.Distinction(), latestRegImage)
		if err != nil {
			return err
		}

		err = a.store.Images().Create(latestImage)
		if err != nil {
			return err
		}

		return nil
	}

	if currentImage != nil && currentImage.Digest != latestRegImageDigest {
		latestImage, err = a.CreateStoreImageFromRegistryImage(latestVP.Distinction(), latestRegImage)
		if err != nil {
			return err
		}

		err = a.store.Images().Create(latestImage)
		if err != nil {
			return err
		}

		currentTag.IsLatest = false
		err = a.store.Images().Update(currentImage)
		if err != nil {
			return err
		}

		return nil
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
