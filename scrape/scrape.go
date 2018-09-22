package scrape

import (
	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/versionparser"
)

type scraper struct {
	store store.Store
}

func (s *scraper) updateImage(image string) error {
	regImg, err := registry.NewImage(image, false)
	if err != nil {
		return err
	}

	regTag, err := regImg.Tag()
	if err != nil {
		return err
	}

	_, err = s.store.FindImageByTag(regImg.Repository.FullName(), regTag)
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
