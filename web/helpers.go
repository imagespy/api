package web

import (
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/versionparser"
	log "github.com/sirupsen/logrus"
)

func findSourceImageOfLayer(sourceImageID int, s store.Store) (*store.Image, []*store.Tag, *store.Image, []*store.Tag, error) {
	imagesClient := s.Images()
	tagsClient := s.Tags()
	sourceImage, err := imagesClient.Get(store.ImageGetOptions{ID: sourceImageID})
	if err != nil {
		log.Errorf("getting source image '%d': %s", sourceImageID, err)
		return nil, nil, nil, nil, err
	}

	sourceImageTags, err := tagsClient.List(store.TagListOptions{ImageID: sourceImage.ID})
	if err != nil {
		log.Errorf("reading tags of source image '%d': %s", sourceImageID, err)
		return nil, nil, nil, nil, err
	}

	latestImage, latestTags, err := findLatestImageOfImage(sourceImage, sourceImageTags, s)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return sourceImage, sourceImageTags, latestImage, latestTags, nil
}

func findLatestImageOfImage(i *store.Image, tags []*store.Tag, s store.Store) (*store.Image, []*store.Tag, error) {
	imagesClient := s.Images()
	var latestImage *store.Image
	var latestVP versionparser.VersionParser
	for _, tag := range tags {
		isLatestTag := true
		li, err := imagesClient.Get(store.ImageGetOptions{
			Name:           i.Name,
			TagDistinction: tag.Distinction,
			TagIsLatest:    &isLatestTag,
		})
		if err != nil {
			log.Errorf("reading latest image of source image '%d': %s", i.ID, err)
			return nil, nil, err
		}

		if latestImage == nil {
			latestImage = li
			latestVP = versionparser.FindForVersion(tag.Name)
		} else {
			vp := versionparser.FindForVersion(tag.Name)
			if li.Digest != latestImage.Digest && latestVP.Weight() < vp.Weight() {
				latestImage = li
				latestVP = vp
			}
		}
	}

	latestTags, err := s.Tags().List(store.TagListOptions{ImageID: latestImage.ID})
	if err != nil {
		log.Errorf("reading tags of latest image of source image '%d': %s", latestImage.ID, err)
		return nil, nil, err
	}

	return latestImage, latestTags, nil
}
