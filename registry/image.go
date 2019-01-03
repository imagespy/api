package registry

import (
	"fmt"

	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	reg "github.com/genuinetools/reg/registry"
)

type image struct {
	parsed        reg.Image
	platforms     []Platform
	populated     bool
	regClient     *reg.Registry
	repository    Repository
	schemaVersion int
}

func (i *image) Digest() (string, error) {
	if i.parsed.Digest.String() == "" {
		err := i.populate()
		if err != nil {
			return "", err
		}
	}

	return i.parsed.Digest.String(), nil
}

func (i *image) Platform(arch string, os string) (Platform, error) {
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

func (i *image) Platforms() ([]Platform, error) {
	if i.populated == false {
		err := i.populate()
		if err != nil {
			return nil, err
		}
	}

	return i.platforms, nil
}

func (i *image) Repository() Repository {
	return i.repository
}

func (i *image) SchemaVersion() (int, error) {
	if i.populated == false {
		err := i.populate()
		if err != nil {
			return 0, err
		}
	}

	return i.schemaVersion, nil
}

func (i *image) Tag() (string, error) {
	return i.parsed.Tag, nil
}

func (i *image) populate() error {
	log.Debug("Populating image")
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
		p := &PlatformV2{
			architecture: "amd64",
			digest:       "",
			image:        i,
			os:           "linux",
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
			digest:    "",
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
