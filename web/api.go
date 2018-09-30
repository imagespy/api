package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/versionparser"
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
	scraper         scrape.Scraper
	Store           store.Store
}

func (h *Handler) Image(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "*")
	regImage, err := registry.NewImage(imageID, false)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tag, err := regImage.Tag()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vp := versionparser.FindForVersion(tag)
	image, err := h.Store.FindImageWithTagsByTag(regImage.Repository.FullName(), tag)
	var latestImage *store.Image
	if err != nil {
		if err == store.ErrDoesNotExist {
			image, _, err = h.Store.CreateImageFromRegistryImage(vp.Distinction(), regImage)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			latestImage, err = scrape.ScrapeLatestImageOfRegistryImage(regImage, h.Store)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		latestImage, err = h.Store.FindLatestImageWithTagsByDistinction(vp.Distinction(), regImage.Repository.FullName())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	b, err := h.imageSerializer(image, latestImage)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
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
		imageSerializer: jsonImageSerializer,
		Store:           store,
	}
	r := chi.NewRouter()
	r.Get("/v2/images/*", h.Image)
	return r
}
