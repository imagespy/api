package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/notifications"
	registryC "github.com/imagespy/registry-client"
	log "github.com/sirupsen/logrus"
)

type registryHandler struct {
	eventDedup      map[string]struct{}
	eventDedupMutex *sync.RWMutex
	regC            *registryC.Registry
	scraper         scrape.Scraper
}

func (rh *registryHandler) registryEvent(w http.ResponseWriter, r *http.Request) {
	log.Debug("processing docker registry event")
	if r.Header.Get("Content-Type") != notifications.EventsMediaType {
		log.Debug("docker registry event contains unsupported content type")
		w.WriteHeader(http.StatusOK)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("reading registry event payload: %s", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	defer r.Body.Close()
	envelope := &notifications.Envelope{}
	err = json.Unmarshal(payload, envelope)
	if err != nil {
		log.Errorf("unmarshalling registry event payload: %s", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	for _, event := range envelope.Events {
		if event.Action != notifications.EventActionPush || event.Target.MediaType != schema2.MediaTypeManifest {
			w.WriteHeader(http.StatusOK)
			return
		}

		targetURL, err := url.ParseRequestURI(event.Target.URL)
		if err != nil {
			log.Error(err)
			continue
		}

		port := ":" + targetURL.Port()
		if port == ":443" || port == ":80" {
			port = ""
		}

		imageName := fmt.Sprintf("%s%s/%s:%s", targetURL.Hostname(), port, event.Target.Repository, event.Target.Tag)
		rh.eventDedupMutex.RLock()
		_, exists := rh.eventDedup[imageName]
		rh.eventDedupMutex.RUnlock()
		if !exists {
			rh.eventDedupMutex.Lock()
			rh.eventDedup[imageName] = struct{}{}
			rh.eventDedupMutex.Unlock()
			go func() {
				imageName := imageName
				defer func() {
					rh.eventDedupMutex.Lock()
					delete(rh.eventDedup, imageName)
					rh.eventDedupMutex.Unlock()
				}()

				domain, path, tag, _, err := registry.ParseImage(imageName)
				if err != nil {
					log.Errorf("parsing name of image '%s': %s", imageName, err)
					return
				}

				regImage := registryC.Image{
					Domain:     domain,
					Repository: path,
					Tag:        tag,
				}

				repo, err := rh.regC.RepositoryFromString(imageName)
				if err != nil {
					log.Errorf("creating repository from string '%s': %s", imageName, err)
					return
				}

				rh.scraper.ScrapeImage(regImage, repo)
				rh.scraper.ScrapeLatestImage(regImage, repo)
			}()
		}
	}

	w.WriteHeader(http.StatusOK)
}
