package scrape

import (
	"database/sql"
	"fmt"
	"os"
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
	layersInDatabase           []*store.Layer
	layerPositionsInDatabase   []*store.LayerPosition
	platformsInDatabase        []*store.Platform
	tagsInDatabase             []*store.Tag
	additionalImagesInRegistry []registry.Image
	expectedImages             []*store.Image
	expectedLayers             []*store.Layer
	expectedLayerPositions     []*store.LayerPosition
	expectedPlatforms          []*store.Platform
	expectedTags               []*store.Tag
}

func executeTest(t *testing.T, tc testcase, funcToTest string) {
	t.Run(tc.name, func(t *testing.T) {
		if os.Getenv("RUN_SCRAPE_TESTS") != "1" {
			t.Skip("Not executing because scrape tests are disabled")
			return
		}

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

		for _, iid := range tc.imagesInDatabase {
			err := s.Images().Create(iid)
			if err != nil {
				t.Fatalf("creating image in database failed: %s", err)
			}
		}

		for _, pid := range tc.platformsInDatabase {
			err := s.Platforms().Create(pid)
			if err != nil {
				t.Fatalf("creating platform in database failed: %s", err)
			}
		}

		for _, lid := range tc.layersInDatabase {
			err := s.Layers().Create(lid)
			if err != nil {
				t.Fatalf("creating layer in database failed: %s", err)
			}
		}

		for _, lpid := range tc.layerPositionsInDatabase {
			err := s.LayerPositions().Create(lpid)
			if err != nil {
				t.Fatalf("creating layer position in database failed: %s", err)
			}
		}

		for _, tid := range tc.tagsInDatabase {
			err := s.Tags().Create(tid)
			if err != nil {
				t.Fatalf("creating tag in database failed: %s", err)
			}
		}

		mockRegistry := registryMock.NewRegistry()
		mockRegistry.AddImage(tc.imageToScrape.(*registryMock.Image))
		for _, i := range tc.additionalImagesInRegistry {
			mockRegistry.AddImage(i.(*registryMock.Image))
		}

		mockTime := &timeMock{}

		scraper := &async{store: s, timeFunc: mockTime.Time}
		switch funcToTest {
		case "ScrapeImage":
			err = scraper.ScrapeImage(tc.imageToScrape)
			require.NoError(t, err)
		case "ScrapeLatestImage":
			err = scraper.ScrapeLatestImage(tc.imageToScrape)
			require.NoError(t, err)
		}

		if len(tc.expectedImages) > 0 {
			storedImages, err := s.Images().List(store.ImageListOptions{})
			require.NoError(t, err)

			sort.Slice(tc.expectedImages, func(i, j int) bool { return tc.expectedImages[i].ID < tc.expectedImages[j].ID })
			sort.Slice(storedImages, func(i, j int) bool { return storedImages[i].ID < storedImages[j].ID })
			assert.EqualValues(t, tc.expectedImages, storedImages)
		}

		if len(tc.expectedPlatforms) > 0 {
			storedPlatforms, err := s.Platforms().List(store.PlatformListOptions{})
			require.NoError(t, err)

			sort.Slice(tc.expectedPlatforms, func(i, j int) bool { return tc.expectedPlatforms[i].ID < tc.expectedPlatforms[j].ID })
			sort.Slice(storedPlatforms, func(i, j int) bool { return storedPlatforms[i].ID < storedPlatforms[j].ID })
			assert.EqualValues(t, tc.expectedPlatforms, storedPlatforms)
		}

		if len(tc.expectedLayers) > 0 {
			storedLayers, err := s.Layers().List(store.LayerListOptions{})
			require.NoError(t, err)

			sort.Slice(tc.expectedLayers, func(i, j int) bool { return tc.expectedLayers[i].ID < tc.expectedLayers[j].ID })
			sort.Slice(storedLayers, func(i, j int) bool { return storedLayers[i].ID < storedLayers[j].ID })
			assert.EqualValues(t, tc.expectedLayers, storedLayers)
		}

		if len(tc.expectedLayerPositions) > 0 {
			storedLayerPositions, err := s.LayerPositions().List(store.LayerPositionListOptions{})
			require.NoError(t, err)

			sort.Slice(tc.expectedLayerPositions, func(i, j int) bool { return tc.expectedLayerPositions[i].ID < tc.expectedLayerPositions[j].ID })
			sort.Slice(storedLayerPositions, func(i, j int) bool { return storedLayerPositions[i].ID < storedLayerPositions[j].ID })
			assert.EqualValues(t, tc.expectedLayerPositions, storedLayerPositions)
		}

		if len(tc.expectedTags) > 0 {
			storedTags, err := s.Tags().List(store.TagListOptions{})
			require.NoError(t, err)

			sort.Slice(tc.expectedTags, func(i, j int) bool { return tc.expectedTags[i].ID < tc.expectedTags[j].ID })
			sort.Slice(storedTags, func(i, j int) bool { return storedTags[i].ID < storedTags[j].ID })
			assert.EqualValues(t, tc.expectedTags, storedTags)
		}
	})
}

func TestAsync_ScrapeImage(t *testing.T) {
	testcases := []testcase{
		{
			name: "When the image has not been scraped it scrapes the image",
			imageToScrape: registryMock.NewImage(
				"ABC",
				"dev.local/unittest",
				[]registry.Platform{
					registryMock.NewPlatform("amd64", "xyz", []string{"l123", "l456"}, "qwz", "linux", time.Date(2018, 10, 24, 1, 0, 0, 0, time.UTC)),
				},
				2,
				"1",
			),
			expectedImages: []*store.Image{
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			expectedLayers: []*store.Layer{
				{Digest: "l123", Model: store.Model{ID: 1}, SourceImageIDs: []int{1}},
				{Digest: "l456", Model: store.Model{ID: 2}, SourceImageIDs: []int{1}},
			},
			expectedLayerPositions: []*store.LayerPosition{
				{LayerID: 1, Model: store.Model{ID: 1}, PlatformID: 1, Position: 0},
				{LayerID: 2, Model: store.Model{ID: 2}, PlatformID: 1, Position: 1},
			},
			expectedPlatforms: []*store.Platform{
				{
					Architecture:   "amd64",
					Created:        time.Date(2018, 10, 24, 1, 0, 0, 0, time.UTC),
					CreatedAt:      time.Date(2018, 10, 26, 3, 0, 0, 0, time.UTC),
					Features:       []*store.Feature{},
					ImageID:        1,
					ManifestDigest: "qwz",
					Model:          store.Model{ID: 1},
					OS:             "linux",
					OSFeatures:     []*store.OSFeature{},
					OSVersion:      "",
					Variant:        "",
				},
			},
			expectedTags: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
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
					CreatedAt:     time.Date(2018, 10, 25, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 25, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			tagsInDatabase: []*store.Tag{
				&store.Tag{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Name: "1"},
			},
			expectedImages: []*store.Image{
				{
					CreatedAt:     time.Date(2018, 10, 25, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			expectedTags: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
				{Distinction: "majorMinor", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 2}, Name: "1.1"},
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
				},
			},
			expectedTags: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: true, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
			},
		},
		{
			name:          "When the image exists but the tag is not flagged as latest it marks the tag as latest",
			imageToScrape: registryMock.NewImage("ABC", "dev.local/unittest", []registry.Platform{}, 2, "1"),
			imagesInDatabase: []*store.Image{
				{
					CreatedAt:     time.Date(2018, 10, 25, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 25, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			tagsInDatabase: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Name: "1"},
			},
			expectedImages: []*store.Image{
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 25, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			expectedTags: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: true, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
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
					CreatedAt:     time.Date(2018, 10, 25, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 25, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			tagsInDatabase: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: true, IsTagged: true, Name: "1"},
			},
			expectedImages: []*store.Image{
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 25, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 25, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "DEF",
					Model:         store.Model{ID: 2},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			expectedTags: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 1}, Name: "1"},
				{Distinction: "major", ImageID: 2, IsLatest: true, IsTagged: true, Model: store.Model{ID: 2}, Name: "2"},
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
					CreatedAt:     time.Date(2018, 10, 25, 1, 0, 0, 0, time.UTC),
					Digest:        "DEF",
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 25, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			tagsInDatabase: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: true, IsTagged: true, Name: "2"},
			},
			expectedImages: []*store.Image{
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 25, 1, 0, 0, 0, time.UTC),
					Digest:        "DEF",
					Model:         store.Model{ID: 1},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 25, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
				&store.Image{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "GHI",
					Model:         store.Model{ID: 2},
					Name:          "dev.local/unittest",
					ScrapedAt:     time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			expectedTags: []*store.Tag{
				{Distinction: "major", ImageID: 1, IsLatest: false, IsTagged: true, Model: store.Model{ID: 1}, Name: "2"},
				{Distinction: "major", ImageID: 2, IsLatest: true, IsTagged: true, Model: store.Model{ID: 2}, Name: "2"},
			},
		},
		{
			name: "When a new image is scraped it sets the IDs of the source images of each layer",
			imageToScrape: registryMock.NewImage(
				"GHI",
				"dev.local/unittest",
				[]registry.Platform{
					registryMock.NewPlatform("amd64", "opq", []string{"l1", "l2"}, "bbb", "linux", time.Date(2018, 10, 24, 1, 0, 0, 0, time.UTC)),
				},
				2,
				"2",
			),
			expectedLayers: []*store.Layer{
				{Digest: "l1", Model: store.Model{ID: 1}, SourceImageIDs: []int{1}},
				{Digest: "l2", Model: store.Model{ID: 2}, SourceImageIDs: []int{1}},
			},
			expectedLayerPositions: []*store.LayerPosition{
				{LayerID: 1, Model: store.Model{ID: 1}, PlatformID: 1, Position: 0},
				{LayerID: 2, Model: store.Model{ID: 2}, PlatformID: 1, Position: 1},
			},
		},
		{
			name: "When a new image is scraped that is the source of already existing layers it sets the IDs of the source images of each layer",
			imagesInDatabase: []*store.Image{
				{
					CreatedAt:     time.Date(2018, 10, 26, 1, 0, 0, 0, time.UTC),
					Digest:        "ABC",
					Name:          "dev.local/derived-image",
					ScrapedAt:     time.Date(2018, 10, 26, 4, 0, 0, 0, time.UTC),
					SchemaVersion: 2,
				},
			},
			layersInDatabase: []*store.Layer{
				{Digest: "l1", SourceImageIDs: []int{1}},
				{Digest: "l2", SourceImageIDs: []int{1}},
				{Digest: "l3", SourceImageIDs: []int{1}},
			},
			layerPositionsInDatabase: []*store.LayerPosition{
				{LayerID: 1, Model: store.Model{ID: 1}, PlatformID: 1, Position: 0},
				{LayerID: 2, Model: store.Model{ID: 2}, PlatformID: 1, Position: 1},
				{LayerID: 3, Model: store.Model{ID: 3}, PlatformID: 1, Position: 2},
			},
			platformsInDatabase: []*store.Platform{
				{
					Created:      time.Date(2018, 10, 26, 2, 0, 0, 0, time.UTC),
					CreatedAt:    time.Date(2018, 10, 26, 3, 0, 0, 0, time.UTC),
					Architecture: "amd64",
					ImageID:      1,
					Model:        store.Model{ID: 1},
					OS:           "linux",
				},
			},
			imageToScrape: registryMock.NewImage(
				"GHI",
				"dev.local/source-image",
				[]registry.Platform{
					registryMock.NewPlatform("amd64", "opq", []string{"l1", "l2"}, "bbb", "linux", time.Date(2018, 10, 24, 1, 0, 0, 0, time.UTC)),
				},
				2,
				"2",
			),
			expectedLayers: []*store.Layer{
				{Digest: "l1", Model: store.Model{ID: 1}, SourceImageIDs: []int{2}},
				{Digest: "l2", Model: store.Model{ID: 2}, SourceImageIDs: []int{2}},
				{Digest: "l3", Model: store.Model{ID: 3}, SourceImageIDs: []int{1}},
			},
			expectedLayerPositions: []*store.LayerPosition{
				{LayerID: 1, Model: store.Model{ID: 1}, PlatformID: 1, Position: 0},
				{LayerID: 2, Model: store.Model{ID: 2}, PlatformID: 1, Position: 1},
				{LayerID: 3, Model: store.Model{ID: 3}, PlatformID: 1, Position: 2},
				{LayerID: 1, Model: store.Model{ID: 4}, PlatformID: 2, Position: 0},
				{LayerID: 2, Model: store.Model{ID: 5}, PlatformID: 2, Position: 1},
			},
		},
	}

	for _, tc := range testcases {
		executeTest(t, tc, "ScrapeLatestImage")
	}
}
