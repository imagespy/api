package registry

import (
	"fmt"

	reg "github.com/genuinetools/reg/registry"
)

type repository struct {
	images      []Image
	initialized bool
	name        string
	regClient   *reg.Registry
}

func (r *repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.regClient.Domain, r.name)
}

func (r *repository) Image(digest string, tag string) Image {
	return r.newImage(digest, tag)
}

func (r *repository) Images() ([]Image, error) {
	if r.initialized {
		return r.images, nil
	}

	log.Debugf("Initializing images in repository %s", r.FullName())
	tags, err := r.regClient.Tags(r.name)
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		r.images = append(r.images, r.newImage("", tag))
	}

	r.initialized = true
	return r.images, nil
}

func (r *repository) newImage(digest string, tag string) Image {
	var suffix string

	if digest != "" {
		suffix = "@" + digest
	} else {
		suffix = ":" + tag
	}

	parsed, _ := reg.ParseImage(r.regClient.Domain + "/" + r.name + suffix)
	return &image{
		parsed:     parsed,
		regClient:  r.regClient,
		repository: r,
	}
}
