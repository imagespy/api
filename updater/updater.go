package updater

import (
	"fmt"
	"sync"

	"github.com/imagespy/api/scrape"

	"github.com/Jeffail/tunny"
	"github.com/imagespy/api/store"
	log "github.com/sirupsen/logrus"
)

type Updater interface {
	Run() error
}

type simpleUpdater struct {
	scraper     scrape.Scraper
	store       store.Store
	workerCount int
}

func (s *simpleUpdater) Run() error {
	b := true
	tags, err := s.store.Tags().List(store.TagListOptions{
		IsLatest: &b,
	})
	if err != nil {
		return fmt.Errorf("simpleUpdater.Run - retrieving tags: %s", err)
	}

	imagesClient := s.store.Images()
	grouped := map[string][]string{}
	for _, tag := range tags {
		image, err := imagesClient.Get(store.ImageGetOptions{
			ID: tag.ImageID,
		})
		if err != nil {
			return fmt.Errorf("simpleUpdater.Run - retrieving image for tag %d: %s", tag.ID, err)
		}

		imgName := image.Name + ":" + tag.Name
		group, exists := grouped[image.Name]
		if exists {
			group = append(group, imgName)
		} else {
			grouped[image.Name] = []string{imgName}
		}
	}

	pool := tunny.NewFunc(s.workerCount, func(payload interface{}) interface{} {
		images, ok := payload.([]string)
		if !ok {
			log.Error("unable to cast payload to []registry.Image")
			return nil
		}

		for _, img := range images {
			err := s.scraper.ScrapeLatestImageByName(img)
			log.Errorf("simpleUpdater.Run - unable to scrape latest image: %s", err)
		}

		return nil
	})
	wg := &sync.WaitGroup{}
	wg.Add(len(grouped))
	for repo, group := range grouped {
		log.Debugf("Dispatching repository %s", repo)
		payload := group
		go func() {
			pool.Process(payload)
			defer wg.Done()
		}()
	}

	wg.Wait()
	return nil
}

func NewSimpleUpdater(scraper scrape.Scraper, s store.Store, wc int) Updater {
	return &simpleUpdater{
		store:       s,
		workerCount: wc,
	}
}
