package registry

import (
	"fmt"

	"github.com/docker/distribution/reference"
)

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
