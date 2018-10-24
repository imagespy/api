package scrape

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	_ "github.com/go-sql-driver/mysql"
	"github.com/imagespy/api/registry"
	registryMock "github.com/imagespy/api/registry/mock"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/store/gorm"
)

func TestAsync_ScrapeLatestImageForTags(t *testing.T) {
	testcases := []struct {
		name             string
		imagesInDatabase *store.Image
		registryImages   []registry.Image
		expectedImages   []*store.Image
	}{
		{
			name: "When no image and tag have been scraped it creates them",
			registryImages: []registry.Image{
				registryMock.NewImage("ABC", "dev.local/unittest", nil, 2, "1"),
			},
			expectedLatestImage: &store.Image{Digest: "ABC", Name: "dev.local/unittest", SchemaVersion: 2},
			expectedLatestTag:   &store.Tag{Distinction: "major", IsLatest: true, IsTagged: true, Name: "1"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			connection := "root:root@tcp(127.0.0.1:33306)/?charset=utf8&parseTime=True&loc=Local"
			db, err := sql.Open("mysql", connection)
			if err != nil {
				t.Fatalf("Unable to connect to database: %s", err)
			}

			_, err = db.Exec("DROP DATABASE IF EXISTS imagespy")
			if err != nil {
				t.Fatalf("Unable to drop database before test: %s", err)
			}

			_, err = db.Exec("CREATE DATABASE imagespy")
			if err != nil {
				t.Fatalf("Unable to create database before test: %s", err)
			}

			db.Close()
			s, err := gorm.New(connection)
			if err != nil {
				t.Fatalf("store unable to connect to database: %s", err)
			}

			err = s.Migrate()
			if err != nil {
				t.Fatalf("migration falied: %s", err)
			}

			mockRegistry := registryMock.NewRegistry()
			for _, i := range tc.registryImages {
				mockRegistry.AddImage(i.(*registryMock.Image))
			}

			scraper := &async{reg: mockRegistry, store: s}
			scraper.ScrapeLatestImageForImages([]string{"dev.local/unittest:test"})

			storedImages, err := s.Images().List(store.ImageListOptions{})
			require.NoError(t, err)

			assert.ObjectsAreEqual(tc.expectedImages, storedImages)
		})
	}
}
