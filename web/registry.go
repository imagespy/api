package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/imagespy/api/scrape"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/notifications"
	log "github.com/sirupsen/logrus"
)

type registryHandler struct {
	eventDedup      map[string]struct{}
	eventDedupMutex *sync.RWMutex
	scraper         scrape.Scraper
}

func (rh *registryHandler) registryEvent(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != notifications.EventsMediaType {
		w.WriteHeader(http.StatusOK)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Unable to read registry event paylod: %s", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	defer r.Body.Close()
	envelope := &notifications.Envelope{}
	err = json.Unmarshal(payload, envelope)
	if err != nil {
		log.Errorf("Unable to unmarshal registry event paylod: %s", err)
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

		imageName := fmt.Sprintf("%s:%s/%s:%s", targetURL.Hostname(), targetURL.Port(), event.Target.Repository, event.Target.Tag)
		rh.eventDedupMutex.RLock()
		_, exists := rh.eventDedup[imageName]
		rh.eventDedupMutex.RUnlock()
		if !exists {
			rh.eventDedupMutex.Lock()
			rh.eventDedup[imageName] = struct{}{}
			rh.eventDedupMutex.Unlock()
			go func() {
				defer func() {
					rh.eventDedupMutex.Lock()
					delete(rh.eventDedup, imageName)
					rh.eventDedupMutex.Unlock()
				}()

				rh.scraper.ScrapeImageByName(imageName)
				rh.scraper.ScrapeLatestImageByName(imageName)
			}()
		}
	}
}
