package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/imagespy/api/store"
	log "github.com/sirupsen/logrus"
)

type layerSerialize struct {
	Digest       string            `json:"digest"`
	SourceImages []*imageSerialize `json:"source_images"`
}

type layersHandler struct {
	serializer func(interface{}) ([]byte, error)
	store      store.Store
}

func (h *layersHandler) layers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	digestInput := vars["digest"]
	layer, err := h.store.Layers().Get(store.LayerGetOptions{Digest: digestInput})
	if err != nil {
		if err == store.ErrDoesNotExist {
			log.Infof("layer %s does not exist", digestInput)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Errorf("reading layer: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serialization := &layerSerialize{Digest: layer.Digest}
	for _, sourceImageID := range layer.SourceImageIDs {
		sourceImage, sourceImageTags, latestImage, latestTags, err := findSourceImageOfLayer(sourceImageID, h.store)
		if err != nil {
			log.Errorf("layersHandler.layers: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		imageSerialization := convertImageToResult(sourceImage, sourceImageTags, latestImage, latestTags)
		serialization.SourceImages = append(serialization.SourceImages, imageSerialization)
	}

	b, err := h.serializer(serialization)
	if err != nil {
		log.Errorf("serializing layer '%s': %s", digestInput, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	addCacheHeaders(w)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
