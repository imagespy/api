package updater

import (
	"testing"

	"github.com/golang/mock/gomock"
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
		Return(&store.Image{Digest: "abc", Name: "first"}, nil).
		AnyTimes()
	imageStore.EXPECT().
		Get(gomock.Eq(store.ImageGetOptions{ID: 2})).
		Return(&store.Image{Digest: "def", Name: "second"}, nil).
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

	scraper := scrape.NewMockScraper(ctrl)
	scraper.EXPECT().
		ScrapeLatestImageByName(gomock.Eq("first:1")).
		Return(nil)
	scraper.EXPECT().
		ScrapeLatestImageByName(gomock.Eq("first:latest")).
		Return(nil)
	scraper.EXPECT().
		ScrapeLatestImageByName(gomock.Eq("second:v3")).
		Return(nil)

	s := &groupingUpdater{
		scraper: scraper,
		store:   store,
	}
	s.dispatchFunc = func(groups map[string][]string) {
		for _, group := range groups {
			s.processRepository(group)
		}
	}

	err := s.Run()
	assert.NoError(t, err)
}
