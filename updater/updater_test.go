package updater

import (
	"testing"

	"github.com/golang/mock/gomock"
	registryMock "github.com/imagespy/api/registry/mock"
	"github.com/imagespy/api/scrape"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/store/mock"
	"github.com/stretchr/testify/assert"
)

func TestSimpleUpdater_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	imageStore := mock.NewMockImageStore(ctrl)
	imageStore.EXPECT().
		Get(gomock.Eq(store.ImageGetOptions{ID: 1})).
		Return(&store.Image{Digest: "abc", Name: "unit.test/first"}, nil).
		AnyTimes()
	imageStore.EXPECT().
		Get(gomock.Eq(store.ImageGetOptions{ID: 2})).
		Return(&store.Image{Digest: "def", Name: "unit.test/second"}, nil).
		AnyTimes()

	tagStore := mock.NewMockTagStore(ctrl)
	b := true
	tagStore.EXPECT().
		List(gomock.Eq(store.TagListOptions{IsLatest: &b})).
		Return([]*store.Tag{
			{Distinction: "major", ImageID: 1, IsLatest: true, IsTagged: true, Name: "1"},
			{Distinction: "static", ImageID: 1, IsLatest: true, IsTagged: true, Name: "latest"},
			{Distinction: "major", ImageID: 2, IsLatest: true, IsTagged: true, Name: "v3"},
		}, nil)

	store := mock.NewMockStore(ctrl)
	store.EXPECT().
		Images().
		Return(imageStore).
		AnyTimes()
	store.EXPECT().
		Tags().
		Return(tagStore).
		AnyTimes()

	rmi1 := registryMock.NewImage("", "unit.test/first", nil, 2, "1")
	rmi2 := registryMock.NewImage("", "unit.test/first", nil, 2, "latest")
	rmi3 := registryMock.NewImage("", "unit.test/second", nil, 2, "v3")
	rm := registryMock.NewRegistry()
	rm.AddImage(rmi1)
	rm.AddImage(rmi2)
	rm.AddImage(rmi3)

	scraper := scrape.NewMockScraper(ctrl)
	scraper.EXPECT().
		ScrapeLatestImage(rmi1).
		Return(nil)
	scraper.EXPECT().
		ScrapeLatestImage(rmi2).
		Return(nil)
	scraper.EXPECT().
		ScrapeLatestImage(rmi3).
		Return(nil)

	s := &groupingUpdater{
		registry: rm,
		scraper:  scraper,
		store:    store,
	}
	s.dispatchFunc = func(groups map[string][]string) {
		for _, group := range groups {
			s.processRepository(group)
		}
	}

	err := s.Run()
	assert.NoError(t, err)
}
