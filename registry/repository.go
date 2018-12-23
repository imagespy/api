package registry

import (
	"fmt"
	"log"

	reg "github.com/genuinetools/reg/registry"
)

type repository struct {
	name      string
	regClient *reg.Registry
}

func (r *repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.regClient.Domain, r.name)
}

func (r *repository) Image(digest string, tag string) Image {
	return r.newImage(digest, tag)
}

func (r *repository) Images() ([]Image, error) {
	images := []Image{}
	tags, err := r.regClient.Tags(r.name)
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		images = append(images, r.newImage("", tag))
	}

	return images, nil
}

func (r *repository) newImage(digest string, tag string) Image {
	var suffix string

	if digest != "" {
		suffix = "@" + digest
	} else {
		suffix = ":" + tag
	}

	parsed, err := reg.ParseImage(r.name + suffix)
	if err != nil {
		log.Fatal(err)
	}

	return &image{
		parsed:     parsed,
		regClient:  r.regClient,
		repository: r,
	}
}
