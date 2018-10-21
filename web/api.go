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
	imageSerializer func(image *store.Image, latestImage *store.Image) ([]byte, error)
	Store           store.Store
	eventDedup      map[string]struct{}
	eventDedupMutex *sync.RWMutex
}

func (h *Handler) Image(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "*")
	image, latestImage, err := h.scrapeImageAndLatestImage(imageID)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b, err := h.imageSerializer(image, latestImage)
	if err != nil {
		log.Error(err)
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

				_, _, err := h.scrapeImageAndLatestImage(imageName)
				if err != nil {
					log.Error(err)
				}
			}()
		}
	}
}

func (h *Handler) scrapeImageAndLatestImage(imageName string) (*store.Image, *store.Image, error) {
	log.Debugf("Scraping image '%s'", imageName)
	regImage, err := registry.NewImage(imageName, registry.Opts{Insecure: true})
	if err != nil {
		return nil, nil, err
	}

	tag, err := regImage.Tag()
	if err != nil {
		return nil, nil, err
	}

	images, err := h.Store.Images().List(store.ImageListOptions{
		Name:    regImage.Repository().FullName(),
		TagName: tag,
	})
	if err != nil {
		return nil, nil, err
	}

	vp := versionparser.FindForVersion(tag)
	var image *store.Image
	if len(images) == 0 {
		digest, err := regImage.Digest()
		if err != nil {
			return nil, nil, err
		}

		image, err = h.Store.Images().Get(digest)
		if err != nil {
			if err == store.ErrDoesNotExist {
				log.Debug("Creating new image")
				image, err = scrape.CreateStoreImageFromRegistryImage(vp.Distinction(), regImage)
				if err != nil {
					return nil, nil, err
				}

				err = h.Store.Images().Create(image)
				if err != nil {
					return nil, nil, err
				}
			} else {
				return nil, nil, err
			}
		}

		_, err = scrape.ScrapeLatestImageOfRegistryImage(regImage, h.Store)
		if err != nil {
			return nil, nil, err
		}
	} else {
		log.Debug("Updating existing image")
		image = images[0]
		t := &store.Tag{
			Distinction: vp.Distinction(),
			ImageID:     image.ID,
			IsLatest:    false,
			IsTagged:    true,
			Name:        tag,
		}
		image.Tags = append(image.Tags, t)
		err = h.Store.Images().Update(image)
		if err != nil {
			return nil, nil, err
		}

		_, err = scrape.ScrapeLatestImageOfRegistryImage(regImage, h.Store)
		if err != nil {
			return nil, nil, err
		}
	}

	fmt.Printf("%#v\n", image)

	b := true
	latestImages, err := h.Store.Images().List(store.ImageListOptions{
		Name:           regImage.Repository().FullName(),
		TagIsLatest:    &b,
		TagDistinction: vp.Distinction(),
	})
	if err != nil {
		return nil, nil, err
	}

	if len(latestImages) == 0 {
		return nil, nil, fmt.Errorf("no latest image found")
	}

	return image, latestImages[0], nil
}

func jsonImageSerializer(image *store.Image, latestImage *store.Image) ([]byte, error) {
	imageSerialized := &imageSerialize{
		Digest: image.Digest,
		Name:   image.Name,
	}
	for _, tag := range image.Tags {
		imageSerialized.Tags = append(imageSerialized.Tags, tag.Name)
	}

	latestImageSerialized := &latestImageSerialize{
		Digest: latestImage.Digest,
		Name:   latestImage.Name,
	}
	for _, latestTag := range latestImage.Tags {
		latestImageSerialized.Tags = append(latestImageSerialized.Tags, latestTag.Name)
	}

	imageSerialized.LatestImage = latestImageSerialized
	b, err := json.Marshal(imageSerialized)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

func Init(store store.Store) http.Handler {
	h := &Handler{
		eventDedup:      map[string]struct{}{},
		eventDedupMutex: &sync.RWMutex{},
		imageSerializer: jsonImageSerializer,
		Store:           store,
	}
	r := chi.NewRouter()
	r.Get("/v2/images/*", h.Image)
	r.Post("/event", h.registryEvent)
	return r
}
