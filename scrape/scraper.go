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
	return &async{reg: reg, store: s}
}

type async struct {
	reg   registry.Registry
	store store.Store
}

func (a *async) ScrapeLatestImageForImages(imageNames []string) {
	for _, name := range imageNames {
		regImg, err := a.reg.Image(name)
		if err != nil {
			log.Errorf("retrieve image %s: %s", name, err)
			continue
		}

		_, err = ScrapeLatestImageOfRegistryImage(regImg, a.store)
		if err != nil {
			log.Error(err)
			continue
		}
	}
}

func (a *async) ScrapeImage(repository, tagRef string) (*store.Image, error) {
	if tagRef == "" {
		return nil, fmt.Errorf("Variable tagRef is empty")
	}

	imageClient := a.store.Images()
	imageList, err := imageClient.List(store.ImageListOptions{
		Name:    repository,
		TagName: tagRef,
	})
	if err != nil {
		return nil, err
	}

	if len(imageList) > 0 {
		return nil, fmt.Errorf("Expected only one tagged image for %s:%s, got %d", repository, tagRef, len(imageList))
	}

	vp := versionparser.FindForVersion(tagRef)
	regImage, err := a.reg.Image(fmt.Sprintf("%s:%s", repository, tagRef))
	if err != nil {
		return nil, err
	}
	if len(imageList) == 0 {
		image, err := CreateStoreImageFromRegistryImage(vp.Distinction(), regImage)
		if err != nil {
			return nil, err
		}

		err = imageClient.Create(image)
		if err != nil {
			return nil, err
		}

		return image, nil
	}

	image := imageList[0]
	tag := &store.Tag{
		Distinction: vp.Distinction(),
		ImageID:     image.ID,
		IsLatest:    false,
		IsTagged:    true,
		Name:        tagRef,
	}
	image.Tags = append(image.Tags, tag)
	err = imageClient.Update(image)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func ScrapeLatestImageOfRegistryImage(i registry.Image, s store.Store) (*store.Image, error) {
	regImgTag, err := i.Tag()
	if err != nil {
		return nil, err
	}

	b := true
	images, err := s.Images().List(store.ImageListOptions{
		Name:        i.Repository().FullName(),
		TagIsLatest: &b,
		TagName:     regImgTag,
	})
	if err != nil {
		return nil, err
	}

	if len(images) > 1 {
		return nil, fmt.Errorf("query for latest tag of %s:%s returned ambiguous result: expected 1 got %d", i.Repository().FullName(), regImgTag, len(images))
	}

	var currentImage *store.Image
	var currentTag *store.Tag
	if len(images) != 0 {
		currentImage = images[0]
		currentTag, err = currentImage.FindTag(regImgTag)
		if err != nil {
			return nil, err
		}
	}

	regImages, err := i.Repository().Images()
	if err != nil {
		return nil, err
	}

	latestVP := versionparser.FindForVersion(regImgTag)
	latestRegImage := i
	for _, regImageItem := range regImages {
		currentImageTag, err := regImageItem.Tag()
		if err != nil {
			return nil, err
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
		return nil, err
	}

	var latestImage *store.Image
	if currentImage == nil {
		latestImage, err = CreateStoreImageFromRegistryImage(latestVP.Distinction(), latestRegImage)
		if err != nil {
			return nil, err
		}

		err = s.Images().Create(latestImage)
		if err != nil {
			return nil, err
		}

		return latestImage, nil
	}

	if currentImage != nil && currentImage.Digest != latestRegImageDigest {
		latestImage, err = CreateStoreImageFromRegistryImage(latestVP.Distinction(), latestRegImage)
		if err != nil {
			return nil, err
		}

		err = s.Images().Create(latestImage)
		if err != nil {
			return nil, err
		}

		currentTag.IsLatest = false
		err = s.Images().Update(currentImage)
		if err != nil {
			return nil, err
		}

		return latestImage, nil
	}

	return currentImage, nil
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

func CreateStoreImageFromRegistryImage(distinction string, regImg registry.Image) (*store.Image, error) {
	imageDigest, err := regImg.Digest()
	if err != nil {
		return nil, err
	}

	imageSV, err := regImg.SchemaVersion()
	if err != nil {
		return nil, err
	}

	image := &store.Image{
		CreatedAt:     time.Now().UTC(),
		Digest:        imageDigest,
		Name:          regImg.Repository().FullName(),
		SchemaVersion: imageSV,
		ScrapedAt:     time.Now().UTC(),
	}
	tagName, err := regImg.Tag()
	if err != nil {
		return nil, err
	}

	tag := &store.Tag{
		Distinction: distinction,
		IsLatest:    true,
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
			CreatedAt:      time.Now().UTC(),
			Created:        time.Now().UTC(),
			ImageID:        image.ID,
			ManifestDigest: regManifestConfig.Digest().String(),
			OS:             p.OS(),
			OSVersion:      p.OSVersion(),
			Variant:        p.Variant(),
		}

		for _, name := range p.Features() {
			platform.Features = append(platform.Features, &store.Feature{
				CreatedAt: time.Now().UTC(),
				Name:      name,
			})
		}

		for _, name := range p.OSFeatures() {
			platform.OSFeatures = append(platform.OSFeatures, &store.OSFeature{
				CreatedAt: time.Now().UTC(),
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
