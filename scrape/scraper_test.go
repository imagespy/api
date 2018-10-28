package scrape

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	_ "github.com/go-sql-driver/mysql"
	"github.com/imagespy/api/registry"
	registryMock "github.com/imagespy/api/registry/mock"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/store/gorm"
)

type timeMock struct {
	callCount int
}

func (m *timeMock) Time() time.Time {
	m.callCount++
	t, _ := time.Parse("2006-01-02 15:04", fmt.Sprintf("2018-10-26 %02d:00", m.callCount))
	return t.UTC()
}

func TestAsync_ScrapeImage(t *testing.T) {
	testcases := []struct {
		name             string
		repository       string
		tag              string
		imagesInDatabase []*store.Image
		registryImages   []registry.Image
		expectedImages   []*store.Image
	}{
		{
			name:       "When the image has not been scraped it scrapes the image",
			repository: "dev.local/unittest",
			tag:        "1",
			registryImages: []registry.Image{
				registryMock.NewImage("ABC", "dev.local/unittest", []registry.Platform{}, 2, "1"),
			},
			expectedImages: []*store.Image{
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			connection := "root:root@tcp(127.0.0.1:33306)/imagespy?charset=utf8&parseTime=True&loc=UTC"
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

			mig, err := migrate.New("file://../store/gorm/migrations", "mysql://"+connection+"&multiStatements=true")
			if err != nil {
				t.Fatalf("unable to create migration object: %s", err)
			}

			defer mig.Close()
			err = mig.Up()
			if err != nil {
				t.Fatalf("running migrations failed: %s", err)
			}

			s, err := gorm.New(connection)
			if err != nil {
				t.Fatalf("store unable to connect to database: %s", err)
			}
			defer s.Close()

			for _, iid := range tc.imagesInDatabase {
				err := s.Images().Create(iid)
				if err != nil {
					t.Fatalf("creating image in database failed: %s", err)
				}
			}

			mockRegistry := registryMock.NewRegistry()
			for _, i := range tc.registryImages {
				mockRegistry.AddImage(i.(*registryMock.Image))
			}

			mockTime := &timeMock{}

			scraper := &async{reg: mockRegistry, store: s, timeFunc: mockTime.Time}
			_, err = scraper.ScrapeImage(tc.repository, tc.tag)
			require.NoError(t, err)

			storedImages, err := s.Images().List(store.ImageListOptions{})
			require.NoError(t, err)

			assert.EqualValues(t, tc.expectedImages, storedImages)
		})
	}
}
