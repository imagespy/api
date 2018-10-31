package scrape

import (
	"database/sql"
	"fmt"
	"sort"
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

type testcase struct {
	name                       string
	imageToScrape              registry.Image
	imagesInDatabase           []*store.Image
	additionalImagesInRegistry []registry.Image
	expectedImages             []*store.Image
}

func executeTest(t *testing.T, tc testcase, funcToTest string) {
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

		err = mig.Up()
		if err != nil {
			mig.Close()
			t.Fatalf("running migrations failed: %s", err)
		}

		mig.Close()
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
		mockRegistry.AddImage(tc.imageToScrape.(*registryMock.Image))
		for _, i := range tc.additionalImagesInRegistry {
			mockRegistry.AddImage(i.(*registryMock.Image))
		}

		mockTime := &timeMock{}

		scraper := &async{reg: mockRegistry, store: s, timeFunc: mockTime.Time}
		switch funcToTest {
		case "ScrapeImage":
			err = scraper.ScrapeImage(tc.imageToScrape)
			require.NoError(t, err)
		case "ScrapeLatestImage":
			err = scraper.ScrapeLatestImage(tc.imageToScrape)
			require.NoError(t, err)
		}

		storedImages, err := s.Images().List(store.ImageListOptions{})
		require.NoError(t, err)

		sort.Slice(tc.expectedImages, func(i, j int) bool { return tc.expectedImages[i].ID < tc.expectedImages[j].ID })
		sort.Slice(storedImages, func(i, j int) bool { return storedImages[i].ID < storedImages[j].ID })
		assert.EqualValues(t, tc.expectedImages, storedImages)
	})
}

func TestAsync_ScrapeImage(t *testing.T) {
	testcases := []testcase{
		{
			name:          "When the image has not been scraped it scrapes the image",
			imageToScrape: registryMock.NewImage("ABC", "dev.local/unittest", []registry.Platform{}, 2, "1"),
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
		{
			name:          "When the image has already been scraped but the tag does not exist it adds the tag",
			imageToScrape: registryMock.NewImage("ABC", "dev.local/unittest", []registry.Platform{}, 2, "1.1"),
			additionalImagesInRegistry: []registry.Image{
				registryMock.NewImage("ABC", "dev.local/unittest", []registry.Platform{}, 2, "1"),
			},
			imagesInDatabase: []*store.Image{
				{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", IsLatest: false, IsTagged: true, Name: "1"},
					},
				},
			},
			expectedImages: []*store.Image{
				{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
						&store.Tag{Distinction: "majorMinor", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 2}, Name: "1.1"},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		executeTest(t, tc, "ScrapeImage")
	}
}

func TestAsync_ScrapeLatestImage(t *testing.T) {
	testcases := []testcase{
		{
			name:          "When there is not latest image it scrapes the image",
			imageToScrape: registryMock.NewImage("ABC", "dev.local/unittest", []registry.Platform{}, 2, "1"),
			expectedImages: []*store.Image{
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", ImageID: 1, IsLatest: true, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
					},
				},
			},
		},
		{
			name:          "When the image exists but the tag is not flagged as latest it marks the tag as latest",
			imageToScrape: registryMock.NewImage("ABC", "dev.local/unittest", []registry.Platform{}, 2, "1"),
			imagesInDatabase: []*store.Image{
				{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", IsLatest: false, IsTagged: true, Name: "1"},
					},
				},
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
						&store.Tag{Distinction: "major", ImageID: 1, IsLatest: true, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
					},
				},
			},
		},
		{
			name:          "When a newer version of an image with a different tag but the same distinction is added it sets the newer version to be the latest verison",
			imageToScrape: registryMock.NewImage("ABC", "dev.local/unittest", []registry.Platform{}, 2, "1"),
			additionalImagesInRegistry: []registry.Image{
				registryMock.NewImage("DEF", "dev.local/unittest", []registry.Platform{}, 2, "2"),
			},
			imagesInDatabase: []*store.Image{
				{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", IsLatest: true, IsTagged: true, Name: "1"},
					},
				},
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
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "DEF",
					Model:         store.Model{ID: 2},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", ImageID: 2, IsLatest: true, IsTagged: true, Model: store.Model{ID: 2}, Name: "2"},
					},
				},
			},
		},
		{
			name:          "When a newer version of an image with the same tag but different digest is added it sets the newer version to be the latest version",
			imageToScrape: registryMock.NewImage("GHI", "dev.local/unittest", []registry.Platform{}, 2, "2"),
			additionalImagesInRegistry: []registry.Image{
				registryMock.NewImage("DEF", "dev.local/unittest", []registry.Platform{}, 2, "2"),
			},
			imagesInDatabase: []*store.Image{
				{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "DEF",
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", IsLatest: true, IsTagged: true, Name: "2"},
					},
				},
			},
			expectedImages: []*store.Image{
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "DEF",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 1}, Name: "2"},
					},
				},
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "GHI",
					Model:         store.Model{ID: 2},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
					Tags: []*store.Tag{
						&store.Tag{Distinction: "major", ImageID: 2, IsLatest: true, IsTagged: true, Model: store.Model{ID: 2}, Name: "2"},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		executeTest(t, tc, "ScrapeLatestImage")
	}
}
