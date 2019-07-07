package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseImage(t *testing.T) {
	testcases := []struct {
		image  string
		domain string
		path   string
		tag    string
		digest string
	}{
		{
			image:  "debian",
			domain: "docker.io",
			path:   "library/debian",
			tag:    "latest",
			digest: "",
		},
		{
			image:  "imagespy/api:1",
			domain: "docker.io",
			path:   "imagespy/api",
			tag:    "1",
			digest: "",
		},
		{
			image:  "imagespy/api@sha256:3b6aaa0901f2c9483c7757e343ec08d7ad3e4520089d0e92fe20db89101244ec",
			domain: "docker.io",
			path:   "imagespy/api",
			tag:    "",
			digest: "sha256:3b6aaa0901f2c9483c7757e343ec08d7ad3e4520089d0e92fe20db89101244ec",
		},
		{
			image:  "registry.private/myapp:1",
			domain: "registry.private",
			path:   "myapp",
			tag:    "1",
			digest: "",
		},
	}

	for _, tc := range testcases {
		resultDomain, resultPath, resultTag, resultDigest, err := ParseImage(tc.image)
		assert.NoError(t, err)
		assert.Equal(t, tc.domain, resultDomain)
		assert.Equal(t, tc.path, resultPath)
		assert.Equal(t, tc.tag, resultTag)
		assert.Equal(t, tc.digest, resultDigest)
	}
}
