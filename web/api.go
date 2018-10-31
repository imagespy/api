package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/imagespy/api/versionparser"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/notifications"
	"github.com/go-chi/chi"
	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	"github.com/imagespy/api/store"
	log "github.com/sirupsen/logrus"
)

type imageSerialize struct {
	Digest      string                `json:"digest"`
	LatestImage *latestImageSerialize `json:"latest_image"`
	Name        string                `json:"name"`
	Tags        []string              `json:"tags"`
}

type latestImageSerialize struct {
	Digest string   `json:"digest"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
}

type Handler struct {
	imageSerializer func(image *store.Image, tags []*store.Tag, latestImage *store.Image, latestTags []*store.Tag) ([]byte, error)
	scraper         scrape.Scraper
	Store           store.Store
	eventDedup      map[string]struct{}
	eventDedupMutex *sync.RWMutex
}

func (h *Handler) Image(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "*")
	address, path, tagInput, _, err := registry.ParseImage(imageID)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	image, err := h.Store.Images().Get(store.ImageGetOptions{
		Name:    address + "/" + path,
		TagName: tagInput,
	})
	if err != nil {
		if err != store.ErrDoesNotExist {
			log.Errorf("reading initial image: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		regImage, err := registry.NewImage(imageID, registry.Opts{})
		if err != nil {
			log.Errorf("instantiating registry image: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = h.scraper.ScrapeImage(regImage)
		if err != nil {
			log.Errorf("scraping registry image: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = h.scraper.ScrapeLatestImage(regImage)
		if err != nil {
			log.Errorf("scraping latest registry image: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	image, err = h.Store.Images().Get(store.ImageGetOptions{
		Name:    address + "/" + path,
		TagName: tagInput,
	})
	if err != nil {
		log.Errorf("reading image again: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tags, err := h.Store.Tags().List(store.TagListOptions{ImageID: image.ID})
	if err != nil {
		log.Errorf("reading tags of current image: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	isLatestTag := true
	latestImage, err := h.Store.Images().Get(store.ImageGetOptions{
		Name:           address + "/" + path,
		TagDistinction: versionparser.FindForVersion(tagInput).Distinction(),
		TagIsLatest:    &isLatestTag,
	})
	if err != nil {
		log.Errorf("reading latest image: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	latestTags, err := h.Store.Tags().List(store.TagListOptions{ImageID: latestImage.ID})
	if err != nil {
		log.Errorf("reading tags of latest image: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := h.imageSerializer(image, tags, latestImage, latestTags)
	if err != nil {
		log.Errorf("serializing image, latest image and tags: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *Handler) registryEvent(w http.ResponseWriter, r *http.Request) {
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
		h.eventDedupMutex.RLock()
		_, exists := h.eventDedup[imageName]
		h.eventDedupMutex.RUnlock()
		if !exists {
			h.eventDedupMutex.Lock()
			h.eventDedup[imageName] = struct{}{}
			h.eventDedupMutex.Unlock()
			go func() {
				defer func() {
					h.eventDedupMutex.Lock()
					delete(h.eventDedup, imageName)
					h.eventDedupMutex.Unlock()
				}()

				// _, _, err := h.scrapeImageAndLatestImage(imageName)
				// if err != nil {
				// 	log.Error(err)
				// }
			}()
		}
	}
}

func jsonImageSerializer(image *store.Image, tags []*store.Tag, latestImage *store.Image, latestTags []*store.Tag) ([]byte, error) {
	imageSerialized := &imageSerialize{
		Digest: image.Digest,
		Name:   image.Name,
	}
	for _, tag := range tags {
		imageSerialized.Tags = append(imageSerialized.Tags, tag.Name)
	}

	latestImageSerialized := &latestImageSerialize{
		Digest: latestImage.Digest,
		Name:   latestImage.Name,
	}
	for _, latestTag := range latestTags {
		latestImageSerialized.Tags = append(latestImageSerialized.Tags, latestTag.Name)
	}

	imageSerialized.LatestImage = latestImageSerialized
	b, err := json.Marshal(imageSerialized)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

func Init(scraper scrape.Scraper, store store.Store) http.Handler {
	h := &Handler{
		eventDedup:      map[string]struct{}{},
		eventDedupMutex: &sync.RWMutex{},
		imageSerializer: jsonImageSerializer,
		scraper:         scraper,
		Store:           store,
	}
	r := chi.NewRouter()
	r.Get("/v2/images/*", h.Image)
	r.Post("/event", h.registryEvent)
	return r
}
