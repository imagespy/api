package updater

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	registryC "github.com/imagespy/registry-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/imagespy/api/store"
	log "github.com/sirupsen/logrus"
)

const (
	prometheusNamespace = "imagespy_updater"
)

var (
	completionTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Name:      "last_completion_timestamp_seconds",
		Help:      "The timestamp of the last completion of a update run, successful or not.",
	})
	duration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Name:      "duration_seconds",
		Help:      "The duration of the last update run in seconds.",
	})
	failCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Name:      "last_scrape_fails",
		Help:      "The number of failed scrapes during the last run.",
	},
	)
)

type Updater interface {
	Run() error
}

type latestImageUpdater struct {
	dispatchFunc func(groups map[string][]string)
	promPusher   *push.Pusher
	regC         *registryC.Registry
	scraper      scrape.Scraper
	store        store.Store
	workerCount  int
}

func (s *latestImageUpdater) Run() error {
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

func (s *latestImageUpdater) processRepository(images []string) {
	repo, err := s.regC.RepositoryFromString(images[0])
	if err != nil {
		log.Errorf("unable to parse repository from image %s: %s", images[0], err)
		failCount.Inc()
		return
	}

	for _, img := range images {
		log.Debugf("scraping latest image for %s", img)
		domain, path, tag, _, err := registry.ParseImage(img)
		if err != nil {
			log.Errorf("unable to scrape latest image of %s: %s", img, err)
			failCount.Inc()
			continue
		}

		regImage := registryC.Image{
			Domain:     domain,
			Repository: path,
			Tag:        tag,
		}
		err = s.scraper.ScrapeLatestImage(regImage, repo)
		if err != nil {
			log.Errorf("unable to scrape latest image of %s: %s", img, err)
			failCount.Inc()
		}
	}
}

func NewLatestImageUpdater(pushgatewayURL string, regC *registryC.Registry, scraper scrape.Scraper, s store.Store, wc int) Updater {
	su := &latestImageUpdater{
		regC:        regC,
		scraper:     scraper,
		store:       s,
		workerCount: wc,
	}

	if pushgatewayURL != "" {
		registry := prometheus.NewRegistry()
		registry.MustRegister(completionTime, duration, failCount)
		su.promPusher = push.New(pushgatewayURL, "imagespy_updater_latest_image").Gatherer(registry)
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

type allImagesUpdater struct {
	db         *sql.DB
	promPusher *push.Pusher
	regC       *registryC.Registry
	scraper    scrape.Scraper
}

func (a *allImagesUpdater) Run() error {
	failCount.Set(0)
	start := time.Now()
	rows, err := a.db.Query("select name from imagespy_image group by name")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return err
		}

		log.Debugf("Updating image %s...", name)
		repository, err := a.regC.RepositoryFromString(name)
		if err != nil {
			return err
		}

		tags, err := repository.Tags().GetAll()
		if err != nil {
			return err
		}

		for _, tag := range tags {
			image, err := repository.Images().GetByTag(tag)
			if err != nil {
				return err
			}

			err = a.scraper.ScrapeImage(image, repository)
			if err != nil {
				log.Error(err)
				failCount.Inc()
				continue
			}

			err = a.scraper.ScrapeLatestImage(image, repository)
			if err != nil {
				log.Error(err)
				failCount.Inc()
				continue
			}
		}
	}

	duration.Set(time.Since(start).Seconds())
	completionTime.SetToCurrentTime()
	if a.promPusher != nil {
		err = a.promPusher.Add()
		if err != nil {
			return err
		}
	}

	return nil
}

func NewAllImagesUpdater(pushgatewayURL string, db *sql.DB, regC *registryC.Registry, s scrape.Scraper) Updater {
	var promPusher *push.Pusher
	if pushgatewayURL != "" {
		registry := prometheus.NewRegistry()
		registry.MustRegister(completionTime, duration, failCount)
		promPusher = push.New(pushgatewayURL, "imagespy_updater_all").Gatherer(registry)
	}

	return &allImagesUpdater{
		db:         db,
		promPusher: promPusher,
		regC:       regC,
		scraper:    s,
	}
}
