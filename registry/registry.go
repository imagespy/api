package registry

//
import (
	"fmt"
	"log"

	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/api/types"
	reg "github.com/genuinetools/reg/registry"
	digest "github.com/opencontainers/go-digest"
)

const digestSHA256GzippedEmptyTar = digest.Digest("sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4")

type Registry struct {
	regClient *reg.Registry
}

func newRegClient(address string, insecure bool) (*reg.Registry, error) {
	if address == "" || address == "docker.io" {
		address = "index.docker.io"
	}
	auth := types.AuthConfig{ServerAddress: address}

	regClient, err := reg.New(auth, reg.Opt{Insecure: insecure, SkipPing: true})
	if err != nil {
		return nil, err
	}

	transport := &AuthTokenTransport{
		Transport: regClient.Client.Transport,
	}
	regClient.Client.Transport = transport
	return regClient, nil
}

func NewRegistry(address string, insecure bool) (*Registry, error) {
	regClient, err := newRegClient(address, insecure)
	if err != nil {
		return nil, err
	}

	r := &Registry{regClient}
	return r, nil
}

func NewRepository(image string, insecure bool) (*Repository, error) {
	img, err := reg.ParseImage(image)
	if err != nil {
		return nil, err
	}

	regClient, err := newRegClient(img.Domain, insecure)
	if err != nil {
		return nil, err
	}

	return &Repository{
		name:      img.Path,
		regClient: regClient,
	}, err
}

func NewImage(image string, insecure bool) (*Image, error) {
	img, err := reg.ParseImage(image)
	if err != nil {
		return nil, err
	}

	regClient, err := newRegClient(img.Domain, insecure)
	if err != nil {
		return nil, err
	}

	return &Image{
		parsed:    img,
		regClient: regClient,
		repository: &Repository{
			name:      img.Path,
			regClient: regClient,
		},
	}, nil
}

func (r *Registry) Repositories() ([]*Repository, error) {
	repositories := []*Repository{}
	catalogItems, err := r.regClient.Catalog("")
	if err != nil {
		return nil, err
	}

	for _, catalogItem := range catalogItems {
		repositories = append(repositories, &Repository{name: catalogItem, regClient: r.regClient})
	}

	return repositories, nil
}

func (r *Registry) Repository(name string) *Repository {
	return &Repository{name: name, regClient: r.regClient}
}

type Repository struct {
	name      string
	regClient *reg.Registry
}

func (r *Repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.regClient.Domain, r.name)
}

func (r *Repository) Image(digest string, tag string) *Image {
	return r.newImage(digest, tag)
}

func (r *Repository) Images() ([]*Image, error) {
	images := []*Image{}
	tags, err := r.regClient.Tags(r.name)
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		images = append(images, r.newImage("", tag))
	}

	return images, nil
}

func (r *Repository) newImage(digest string, tag string) *Image {
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

	return &Image{
		parsed:     parsed,
		regClient:  r.regClient,
		repository: r,
	}
}

type Image struct {
	parsed        reg.Image
	platforms     []Platform
	populated     bool
	regClient     *reg.Registry
	repository    *Repository
	schemaVersion int
}

func (i *Image) Digest() (string, error) {
	if i.parsed.Digest.String() == "" {
		err := i.populate()
		if err != nil {
			return "", err
		}
	}

	return i.parsed.Digest.String(), nil
}

func (i *Image) Platform(arch string, os string) (Platform, error) {
	if i.populated == false {
		err := i.populate()
		if err != nil {
			return nil, err
		}
	}

	for _, p := range i.platforms {
		if p.Architecture() == arch && p.OS() == os {
			return p, nil
		}
	}

	return nil, fmt.Errorf("%s does not support %s/%s", i.parsed.String(), os, arch)
}

func (i *Image) Platforms() ([]Platform, error) {
	if i.populated == false {
		err := i.populate()
		if err != nil {
			return nil, err
		}
	}

	return i.platforms, nil
}

func (i *Image) SchemaVersion() (int, error) {
	if i.populated == false {
		err := i.populate()
		if err != nil {
			return 0, err
		}
	}

	return i.schemaVersion, nil
}

func (i *Image) Tag() (string, error) {
	return i.parsed.Tag, nil
}

func (i *Image) populate() error {
	if i.parsed.Digest.String() == "" {
		d, err := i.regClient.Digest(i.parsed)
		if err != nil {
			return err
		}

		err = i.parsed.WithDigest(d)
		if err != nil {
			return err
		}
	}

	rawManifest, err := i.regClient.Manifest(i.parsed.Path, i.parsed.Tag)
	if err != nil {
		return err
	}

	switch manifest := rawManifest.(type) {
	case *schema2.DeserializedManifest:
		i.schemaVersion = manifest.SchemaVersion
		d, err := i.regClient.Digest(i.parsed)
		if err != nil {
			return err
		}

		p := &PlatformV2{
			architecture: "amd64",
			digest:       d,
			features:     []string{},
			image:        i,
			os:           "linux",
			osFeatures:   []string{},
			osVersion:    "",
			regClient:    i.regClient,
			variant:      "",
		}
		p.manifest = NewManifestV2(manifest.Manifest, p)

		i.platforms = append(i.platforms, p)
	case *manifestlist.DeserializedManifestList:
		i.schemaVersion = manifest.SchemaVersion
		for _, platformManifest := range manifest.Manifests {
			i.platforms = append(i.platforms, &PlatformV2{
				architecture: platformManifest.Platform.Architecture,
				digest:       platformManifest.Digest,
				features:     platformManifest.Platform.Features,
				image:        i,
				os:           platformManifest.Platform.OS,
				osFeatures:   platformManifest.Platform.OSFeatures,
				osVersion:    platformManifest.Platform.OSVersion,
				regClient:    i.regClient,
				variant:      platformManifest.Platform.Variant,
			})
		}
	case *schema1.SignedManifest:
		i.schemaVersion = manifest.SchemaVersion
		d, err := i.regClient.Digest(i.parsed)
		if err != nil {
			return err
		}

		p := &PlatformV1{
			digest:    d,
			image:     i,
			regClient: i.regClient,
		}
		m, err := NewManifestV1(p, manifest, d)
		if err != nil {
			return err
		}

		p.manifest = m
		i.platforms = append(i.platforms, p)
	default:
		return fmt.Errorf("Unknown manifest type")
	}

	i.populated = true
	return nil
}
