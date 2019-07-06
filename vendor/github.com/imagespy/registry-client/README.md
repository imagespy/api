# registry-client

[![Build Status](https://travis-ci.com/imagespy/registry-client.svg?branch=master)](https://travis-ci.com/imagespy/registry-client)
[![GoDoc](https://godoc.org/github.com/imagespy/registry-client?status.svg)](https://godoc.org/github.com/imagespy/registry-client)

A Docker Registry client.

## Usage

```golang
package main

import (
	"fmt"
	"log"

	"github.com/imagespy/registry-client"
)

func main() {
	reg := &registry.Registry{
		Authenticator: registry.NewTokenAuthenticator(),
		Client:        registry.DefaultClient(),
		Domain:        "docker.io",
	}

	repo, err := reg.RepositoryFromString("golang")
	if err != nil {
		log.Fatal(err)
	}

	img, err := repo.Images().GetByTag("1.12.0")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(img.Digest)
}
```
