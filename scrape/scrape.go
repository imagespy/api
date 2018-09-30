package scrape

import (
	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/versionparser"
)

type Scraper struct {
	store store.Store
}

func ScrapeLatestImageOfRegistryImage(i *registry.Image, s store.Store) (*store.Image, error) {
	tag, err := i.Tag()
	if err != nil {
		return nil, err
	}

	regImages, err := i.Repository.Images()
	if err != nil {
		return nil, err
	}

	latestVP := versionparser.FindForVersion(tag)
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

	latestImage, latestTag, err := s.CreateImageFromRegistryImage(latestVP.Distinction(), latestRegImage)
	if err != nil {
		return nil, err
	}

	if latestTag.IsLatest == false {
		latestTag.IsLatest = true
		s.UpdateTag(latestTag)
	}

	return latestImage, nil
}

func (s *Scraper) updateImage(image string) error {
	regImg, err := registry.NewImage(image, false)
	if err != nil {
		return err
	}

	regTag, err := regImg.Tag()
	if err != nil {
		return err
	}

	_, err = s.store.FindImageWithTagsByTag(regImg.Repository.FullName(), regTag)
	if err == nil {
		return nil
	}

	if err != store.ErrDoesNotExist {
		return err
	}

	vp := versionparser.FindForVersion(regTag)
	_, _, err = s.store.CreateImageFromRegistryImage(vp.Distinction(), regImg)
	if err != nil {
		return err
	}

	return nil
}
