package updater

import (
	"fmt"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/imagespy/api/store"
	log "github.com/sirupsen/logrus"
)

var (
	completionTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "imagespy_updater_last_completion_timestamp_seconds",
		Help: "The timestamp of the last completion of a update run, successful or not.",
	})
	duration = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "imagespy_updater_duration_seconds",
		Help: "The duration of the last update run in seconds.",
	})
	failCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "imagespy_updater_last_scrape_fails",
		Help: "The number of failed scrapes during the last run.",
	})
)

type Updater interface {
	Run() error
}

type groupingUpdater struct {
	dispatchFunc func(groups map[string][]string)
	promPusher   *push.Pusher
	registry     registry.Registry
	scraper      scrape.Scraper
	store        store.Store
	workerCount  int
}

func (s *groupingUpdater) Run() error {
	start := time.Now()
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
	duration.Set(time.Since(start).Seconds())
	completionTime.SetToCurrentTime()
	if s.promPusher != nil {
		err = s.promPusher.Add()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *groupingUpdater) processRepository(images []string) {
	repo, err := s.registry.Repository(images[0])
	if err != nil {
		log.Errorf("unable to parse repository from image %s: %s", images[0], err)
		failCount.Inc()
		return
	}

	for _, img := range images {
		log.Debugf("scraping latest image for %s", img)
		_, _, tag, digest, err := registry.ParseImage(img)
		if err != nil {
			log.Errorf("unable to scrape latest image of %s: %s", img, err)
			failCount.Inc()
			continue
		}

		err = s.scraper.ScrapeLatestImage(repo.Image(digest, tag))
		if err != nil {
			log.Errorf("unable to scrape latest image of %s: %s", img, err)
			failCount.Inc()
		}
	}
}

func NewGroupingUpdater(pushgatewayURL string, r registry.Registry, scraper scrape.Scraper, s store.Store, wc int) Updater {
	su := &groupingUpdater{
		registry:    r,
		scraper:     scraper,
		store:       s,
		workerCount: wc,
	}

	if pushgatewayURL != "" {
		registry := prometheus.NewRegistry()
		registry.MustRegister(completionTime, duration, failCount)
		su.promPusher = push.New(pushgatewayURL, "imagespy").Gatherer(registry)
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
