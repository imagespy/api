package updater

import (
	"fmt"
	"sync"

	"github.com/Jeffail/tunny"
	"github.com/imagespy/api/scrape"

	"github.com/imagespy/api/store"
	log "github.com/sirupsen/logrus"
)

type Updater interface {
	Run() error
}

type groupingUpdater struct {
	scraper      scrape.Scraper
	store        store.Store
	dispatchFunc func(groups map[string][]string)
	workerCount  int
}

func (s *groupingUpdater) Run() error {
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
			grouped[image.Name] = append(group, imgName)
		} else {
			grouped[image.Name] = []string{imgName}
		}
	}

	s.dispatchFunc(grouped)
	return nil
}

func (s *groupingUpdater) processRepository(images []string) {
	for _, img := range images {
		err := s.scraper.ScrapeLatestImageByName(img)
		if err != nil {
			log.Errorf("simpleUpdater.Run - unable to scrape latest image: %s", err)
		}
	}
}

func NewGroupingUpdater(scraper scrape.Scraper, s store.Store, wc int) Updater {
	su := &groupingUpdater{
		scraper:     scraper,
		store:       s,
		workerCount: wc,
	}

	pool := tunny.NewFunc(wc, func(payload interface{}) interface{} {
		images, ok := payload.([]string)
		if !ok {
			log.Error("unable to cast payload to []string")
			return nil
		}

		su.processRepository(images)
		return nil
	})

	ap := asyncProcessor{pool: pool}
	su.dispatchFunc = ap.dispatch

	return su
}

type asyncProcessor struct {
	pool *tunny.Pool
}

func (ap *asyncProcessor) dispatch(groups map[string][]string) {
	wg := &sync.WaitGroup{}
	wg.Add(len(groups))
	for _, group := range groups {
		payload := group
		go func() {
			ap.pool.Process(payload)
			wg.Done()
		}()
	}

	wg.Wait()
}
