package registry

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/imagespy/registry-client"
)

type Config struct {
	Address           string
	Authentication    string
	BasicAuthPassword string
	BasicAuthUsername string
	Insecure          bool
	Protocol          string
	Timeout           time.Duration
}

// Builder builds new registries from configs.
type Builder struct {
	Configs []*Config
}

func (rb *Builder) NewRepositoryFromImage(name string) (*registry.Repository, error) {
	domain, path, _, _, err := ParseImage(name)
	if err != nil {
		return nil, err
	}

	config := rb.findConfig(domain)
	if config == nil {
		return nil, fmt.Errorf("no registry with address %s configured", domain)
	}

	var auth registry.Authenticator
	switch config.Authentication {
	case "basic":
		auth = registry.NewBasicAuthenticator(config.BasicAuthUsername, config.BasicAuthPassword)
	case "token":
		auth = registry.NewTokenAuthenticator()
	default:
		auth = registry.NewNullAuthenticator()
	}

	reg := registry.New(registry.Options{
		Authenticator: auth,
		Client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: config.Insecure,
				},
			},
		},
		Domain:   domain,
		Protocol: config.Protocol,
	})

	return reg.Repository(path), nil
}

func (rb *Builder) findConfig(domain string) *Config {
	for _, c := range rb.Configs {
		if c.Address == domain {
			return c
		}
	}

	return nil
}

// ParseImage receives the name of an image and returns its domain, path, tag and digest.
func ParseImage(n string) (string, string, string, string, error) {
	domain := ""
	path := ""
	tag := ""
	digest := ""

	named, err := reference.ParseNormalizedNamed(n)
	if err != nil {
		return domain, path, tag, digest, fmt.Errorf("parsing image %q failed: %v", n, err)
	}

	named = reference.TagNameOnly(named)
	domain = reference.Domain(named)
	path = reference.Path(named)
	if tagged, ok := named.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	if canonical, ok := named.(reference.Canonical); ok {
		digest = canonical.Digest().String()
	}

	return domain, path, tag, digest, nil
}
