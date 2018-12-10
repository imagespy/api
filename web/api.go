package web

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/imagespy/api/versionparser"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"
	"github.com/imagespy/api/store"
	log "github.com/sirupsen/logrus"
)

var (
	promReqCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	promReqDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "method"},
	)
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

type imageHandler struct {
	serializer func(interface{}) ([]byte, error)
	scraper    scrape.Scraper
	Store      store.Store
}

func (h *imageHandler) createImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["name"]
	address, path, tagInput, _, err := registry.ParseImage(imageID)
	if err != nil {
		log.Errorf("parsing image name: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.Store.Images().Get(store.ImageGetOptions{
		Name:    address + "/" + path,
		TagName: tagInput,
	})
	if err == nil {
		log.Warn("image to create already exists")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err != store.ErrDoesNotExist {
		log.Errorf("reading initial image: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.scraper.ScrapeImageByName(imageID)
	if err != nil {
		log.Errorf("scraping registry image: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.scraper.ScrapeLatestImageByName(imageID)
	if err != nil {
		log.Errorf("scraping latest registry image: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *imageHandler) getImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["name"]
	address, path, tagInput, _, err := registry.ParseImage(imageID)
	if err != nil {
		log.Errorf("parsing image name: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	image, err := h.Store.Images().Get(store.ImageGetOptions{
		Name:    address + "/" + path,
		TagName: tagInput,
	})
	if err != nil {
		if err == store.ErrDoesNotExist {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Errorf("reading image: %s", err)
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

func (h *imageHandler) getImageLayers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["name"]
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
		if err == store.ErrDoesNotExist {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Errorf("layersHandler.layers: reading image: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	platform, err := h.Store.Platforms().Get(store.PlatformGetOptions{
		Architecture: getQueryParam(r, "arch", "amd64"),
		ImageID:      image.ID,
		OS:           getQueryParam(r, "os", "linux"),
		OSVersion:    getQueryParam(r, "os_version", ""),
		Variant:      getQueryParam(r, "variant", ""),
	})
	if err != nil {
		if err == store.ErrDoesNotExist {
			log.Info("layersHandler.layers: platform does not exist")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Errorf("layersHandler.layers: reading platform of image '%d': %s", image.ID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	layerPositions, err := h.Store.LayerPositions().List(store.LayerPositionListOptions{PlatformID: platform.ID})
	if err != nil {
		log.Errorf("layersHandler.layers: reading layer position of platform '%d': %s", platform.ID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := []*layerSerialize{}
	layersClient := h.Store.Layers()
	for _, lp := range layerPositions {
		layer, err := layersClient.Get(store.LayerGetOptions{ID: lp.LayerID})
		if err != nil {
			log.Errorf("layersHandler.layers: reading layer of position '%d': %s", lp.ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		serialization := &layerSerialize{Digest: layer.Digest}
		for _, sourceImageID := range layer.SourceImageIDs {
			sourceImage, sourceImageTags, latestImage, latestTags, err := findSourceImageOfLayer(sourceImageID, h.Store)
			if err != nil {
				log.Errorf("layersHandler.layers: reading source image '%d': %s", sourceImageID, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			imageSerialization := convertImageToResult(sourceImage, sourceImageTags, latestImage, latestTags)
			serialization.SourceImages = append(serialization.SourceImages, imageSerialization)
		}

		result = append(result, serialization)
	}

	b, err := h.serializer(result)
	if err != nil {
		log.Errorf("layersHandler.layers: serializing result: %s", err)
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
	h := &imageHandler{
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

	r := mux.NewRouter()
	r.HandleFunc(`/v2/images/{name:[a-zA-Z0-9/.-:_]+}`, wrapPrometheus("/v2/images/{name}", h.createImage)).Methods("POST")
	r.HandleFunc(`/v2/images/{name:[a-zA-Z0-9/.-:_]+}`, wrapPrometheus("/v2/images/{name}", h.getImage)).Methods("GET")
	r.HandleFunc(`/v2/images/{name:[a-zA-Z0-9/.-:_]+}/layers`, wrapPrometheus("/v2/images/{name}/layers", h.getImageLayers)).Methods("GET")
	r.HandleFunc("/v2/layers/{digest}", wrapPrometheus("/v2/layers/{digest}", lh.layers)).Methods("GET")
	r.HandleFunc("/registry/event", wrapPrometheus("/registry/event", rh.registryEvent)).Methods("POST")
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	return r
}

func getQueryParam(r *http.Request, key, defaultVal string) string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}

	return v
}

func wrapPrometheus(name string, h http.HandlerFunc) http.HandlerFunc {
	return promhttp.InstrumentHandlerDuration(promReqDuration.MustCurryWith(prometheus.Labels{"handler": name}),
		promhttp.InstrumentHandlerCounter(promReqCounter, h))
}
