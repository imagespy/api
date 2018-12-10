package web

import (
	"github.com/imagespy/api/store"
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

	isLatestTag := true
	latestImage, err := imagesClient.Get(store.ImageGetOptions{
		Name:        sourceImage.Name,
		TagIsLatest: &isLatestTag,
	})
	if err != nil {
		log.Errorf("reading latest image of source image '%d': %s", sourceImageID, err)
		return nil, nil, nil, nil, err
	}

	latestTags, err := tagsClient.List(store.TagListOptions{ImageID: latestImage.ID})
	if err != nil {
		log.Errorf("reading tags of latest image of source image '%d': %s", sourceImageID, err)
		return nil, nil, nil, nil, err
	}

	return sourceImage, sourceImageTags, latestImage, latestTags, nil
}
