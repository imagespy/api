package registry

import (
	reg "github.com/genuinetools/reg/registry"
)

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
