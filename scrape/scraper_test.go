package scrape

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/store"
)

type mockStore struct {
	store.Store
	image *store.Image
	tag   *store.Tag
}

func (m *mockStore) CreateImageFromRegistryImage(distinction string, regImg registry.Image) (*store.Image, *store.Tag, error) {
	digest, _ := regImg.Digest()
	schemaVersion, _ := regImg.SchemaVersion()
	tag, _ := regImg.Tag()
	m.image = &store.Image{
		Digest:        digest,
		Name:          regImg.Repository().FullName(),
		SchemaVersion: schemaVersion,
	}
	m.tag = &store.Tag{
		Distinction: distinction,
		Image:       m.image,
		IsLatest:    false,
		IsTagged:    true,
		Name:        tag,
	}
	return m.image, m.tag, nil
}

func (m *mockStore) FindLatestTagWithImage(repository, tag string) (*store.Tag, error) {
	if m.tag == nil || m.tag.IsLatest == false {
		return nil, store.ErrDoesNotExist
	}

	return m.tag, nil
}

func (m *mockStore) UpdateTag(t *store.Tag) error {
	m.tag = t
	return nil
}

func TestAsync_ScrapeLatestImageForTags(t *testing.T) {
	testcases := []struct {
		name                       string
		initialCurrentImage        *store.Image
		initialCurrentTag          *store.Tag
		registryImage              *mockRegistryImage
		registryImagesInRepository []registry.Image
		expectedLatestImage        *store.Image
		expectedLatestTag          *store.Tag
	}{
		{
			name:                "When no image and tag have been scraped it creates them",
			registryImage:       &mockRegistryImage{digest: "ABC", name: "dev.local/unittest", schemaVersion: 2, tag: "1"},
			expectedLatestImage: &store.Image{Digest: "ABC", Name: "dev.local/unittest", SchemaVersion: 2},
			expectedLatestTag:   &store.Tag{Distinction: "major", IsLatest: true, IsTagged: true, Name: "1"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := &mockStore{}
			if tc.initialCurrentImage != nil {
				s.image = tc.initialCurrentImage
			}

			if tc.initialCurrentTag != nil {
				s.tag = tc.initialCurrentTag
			}

			tc.registryImage.repository = &mockRegistryRepository{name: tc.registryImage.name, otherImages: tc.registryImagesInRepository}
			scraper := &async{reg: &mockRegistry{image: tc.registryImage}, store: s}
			scraper.ScrapeLatestImageForImages([]string{"dev.local/unittest:test"})

			require.NotNil(t, s.image)

			assert.Equal(t, tc.expectedLatestImage.Digest, s.image.Digest)
			assert.Equal(t, tc.expectedLatestImage.Name, s.image.Name)
			assert.Equal(t, tc.expectedLatestImage.SchemaVersion, s.image.SchemaVersion)

			assert.Equal(t, tc.expectedLatestTag.Distinction, s.tag.Distinction)
			assert.Equal(t, tc.expectedLatestTag.IsLatest, s.tag.IsLatest)
			assert.Equal(t, tc.expectedLatestTag.IsTagged, s.tag.IsTagged)
			assert.Equal(t, tc.expectedLatestTag.Name, s.tag.Name)
		})
	}
}
