package registry

import (
	"github.com/docker/docker/api/types"
	reg "github.com/genuinetools/reg/registry"
	digest "github.com/opencontainers/go-digest"
)

const digestSHA256GzippedEmptyTar = digest.Digest("sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4")

type Opts struct {
	Insecure bool
	Password string
	Username string
}

type registry struct {
	address   string
	opts      Opts
	regClient *reg.Registry
}

func newRegClient(address string, o Opts) (*reg.Registry, error) {
	if address == "" || address == "docker.io" {
		address = "index.docker.io"
	}
	auth := types.AuthConfig{
		Password:      o.Password,
		ServerAddress: address,
		Username:      o.Username,
	}

	regClient, err := reg.New(auth, reg.Opt{Insecure: o.Insecure, SkipPing: true})
	if err != nil {
		return nil, err
	}

	transport := &AuthTokenTransport{
		Transport: regClient.Client.Transport,
	}
	regClient.Client.Transport = transport
	return regClient, nil
}

func NewRegistry(address string, o Opts) (Registry, error) {
	regClient, err := newRegClient(address, o)
	if err != nil {
		return nil, err
	}

	r := &registry{
		address:   address,
		regClient: regClient,
		opts:      o,
	}
	return r, nil
}

func NewRepository(image string, o Opts) (Repository, error) {
	img, err := reg.ParseImage(image)
	if err != nil {
		return nil, err
	}

	r, err := NewRegistry(img.Domain, o)
	if err != nil {
		return nil, err
	}

	return r.Repository(image)
}

func NewImage(image string, o Opts) (Image, error) {
	img, err := reg.ParseImage(image)
	if err != nil {
		return nil, err
	}

	r, err := NewRegistry(img.Domain, o)
	if err != nil {
		return nil, err
	}

	return r.Image(image)
}

func (r *registry) Address() string {
	return r.address
}

func (r *registry) Image(imageName string) (Image, error) {
	repo, err := r.Repository(imageName)
	if err != nil {
		return nil, err
	}

	img, err := reg.ParseImage(imageName)
	if err != nil {
		return nil, err
	}

	return repo.Image(img.Digest.String(), img.Tag), nil
}

func (r *registry) Repositories() ([]Repository, error) {
	repositories := []Repository{}
	catalogItems, err := r.regClient.Catalog("")
	if err != nil {
		return nil, err
	}

	for _, catalogItem := range catalogItems {
		repo, err := r.Repository(catalogItem)
		if err != nil {
			return nil, err
		}

		repositories = append(repositories, repo)
	}

	return repositories, nil
}

func (r *registry) Repository(imageName string) (Repository, error) {
	img, err := reg.ParseImage(imageName)
	if err != nil {
		return nil, err
	}

	client, err := newRegClient(img.Domain, r.opts)
	if err != nil {
		return nil, err
	}

	return &repository{
		name:      img.Path,
		regClient: client,
	}, nil
}

func ParseImage(n string) (string, string, string, string, error) {
	i, err := reg.ParseImage(n)
	if err != nil {
		return "", "", "", "", err
	}

	var domain string
	if i.Domain == "docker.io" {
		domain = "index.docker.io"
	} else {
		domain = i.Domain
	}

	return domain, i.Path, i.Tag, i.Digest.String(), nil
}
