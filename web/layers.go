package web

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/imagespy/api/store"
	log "github.com/sirupsen/logrus"
)

type layerSerialize struct {
	Digest       string
	SourceImages []*imageSerialize `json:"source_images"`
}

type layersHandler struct {
	serializer func(interface{}) ([]byte, error)
	store      store.Store
}

func (h *layersHandler) layers(w http.ResponseWriter, r *http.Request) {
	digestInput := chi.URLParam(r, "digest")
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

	imagesClient := h.store.Images()
	tagsClient := h.store.Tags()
	serialization := &layerSerialize{Digest: layer.Digest}
	for _, sourceImageID := range layer.SourceImageIDs {
		sourceImage, err := imagesClient.Get(store.ImageGetOptions{ID: sourceImageID})
		if err != nil {
			log.Errorf("getting source image '%d': %s", sourceImageID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		sourceImageTags, err := tagsClient.List(store.TagListOptions{ImageID: sourceImage.ID})
		if err != nil {
			log.Errorf("reading tags of source image '%d': %s", sourceImageID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		isLatestTag := true
		latestImage, err := imagesClient.Get(store.ImageGetOptions{
			Name:        sourceImage.Name,
			TagIsLatest: &isLatestTag,
		})
		if err != nil {
			log.Errorf("reading latest image of source image '%d': %s", sourceImageID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		latestTags, err := tagsClient.List(store.TagListOptions{ImageID: latestImage.ID})
		if err != nil {
			log.Errorf("reading tags of latest image of source image '%d': %s", sourceImageID, err)
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
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
