package web

import (
	"testing"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/store"
)

func TestHandler_Image(t *testing.T) {
	testcases := []struct {
		name                       string
		existingImagesInStore      []*store.Image
		registryImagesInRepository []registry.Image
		expectedOuput              *imageSerialize
	}{
		{
			name: "When no image and tag have been scraped it creates them",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
		})
	}
}
