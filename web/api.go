package web

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/imagespy/api/versionparser"

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
	serializer func(interface{}) ([]byte, error)
	scraper    scrape.Scraper
	Store      store.Store
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

	serialization := convertImageToResult(image, tags, latestImage, latestTags)
	b, err := h.serializer(serialization)
	if err != nil {
		log.Errorf("serializing image, latest image and tags: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func convertImageToResult(image *store.Image, tags []*store.Tag, latestImage *store.Image, latestTags []*store.Tag) *imageSerialize {
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
	return imageSerialized
}

func Init(scraper scrape.Scraper, store store.Store) http.Handler {
	h := &Handler{
		serializer: json.Marshal,
		scraper:    scraper,
		Store:      store,
	}

	rh := &registryHandler{
		eventDedup:      map[string]struct{}{},
		eventDedupMutex: &sync.RWMutex{},
		scraper:         scraper,
	}

	lh := &layersHandler{
		serializer: json.Marshal,
		store:      store,
	}

	r := chi.NewRouter()
	r.Get("/v2/images/*", h.Image)
	r.Get("/v2/layers/{digest}", lh.layers)
	r.Post("/registry/event", rh.registryEvent)
	return r
}
