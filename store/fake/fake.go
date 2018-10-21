package fake

import (
	"encoding/json"

	"github.com/imagespy/api/store"
)

type fakeStore struct {
	is *imageStore
}

func (fs *fakeStore) Images() store.ImageStore {
	return nil
}

func NewStore() store.Store {
	return &fakeStore{is: &imageStore{}}
}

type imageStore struct {
	images []*store.Image
}

func (is *imageStore) Create(i *store.Image) error {
	is.images = append(is.images, i)
	return nil
}

func (is *imageStore) Get(digest string) (*store.Image, error) {
	for _, i := range is.images {
		if i.Digest == digest {
			return i, nil
		}
	}

	return nil, store.ErrDoesNotExist
}

func (is *imageStore) List(o store.ImageListOptions) ([]*store.Image, error) {
	result := is.images
	if o.Digest != "" {
		tmp := []*store.Image{}
		for _, i := range result {
			if i.Digest == o.Digest {
				tmp = append(tmp, i)
			}
		}

		result = tmp
	}

	if o.Name != "" {
		tmp := []*store.Image{}
		for _, i := range result {
			if i.Name == o.Name {
				tmp = append(tmp, i)
			}
		}

		result = tmp
	}

	if o.TagDistinction != "" {
		tmp := []*store.Image{}
		for _, i := range result {
			for _, t := range i.Tags {
				if t.Distinction == o.TagDistinction {
					tmp = append(tmp, i)
				}
			}
		}

		result = tmp
	}

	if o.TagIsLatest != nil {
		tmp := []*store.Image{}
		for _, i := range result {
			for _, t := range i.Tags {
				if t.IsLatest == *o.TagIsLatest {
					tmp = append(tmp, i)
				}
			}
		}

		result = tmp
	}

	if o.TagName != "" {
		tmp := []*store.Image{}
		for _, i := range result {
			for _, t := range i.Tags {
				if t.Name == o.TagName {
					tmp = append(tmp, i)
				}
			}
		}

		result = tmp
	}

	return result, nil
}

func (is *imageStore) Update(i *store.Image) error {
	copy := &store.Image{}
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, copy)
	if err != nil {
		return err
	}

	for index, i := range is.images {
		if i.Digest == copy.Digest {
			is.images[index] = copy
			return nil
		}
	}

	return store.ErrDoesNotExist
}
