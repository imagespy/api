package registry

import (
	"fmt"
	"net/http"
)

// AuthTokenTransport defines the data structure for custom http.Request options.
type AuthTokenTransport struct {
	Transport http.RoundTripper
	authToken string
}

// RoundTrip defines the round tripper for the error transport.
func (t *AuthTokenTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if t.authToken != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.authToken))
	}

	resp, err := t.Transport.RoundTrip(request)
	if err == nil {
		authToken := resp.Header.Get("request-token")
		if authToken != "" {
			t.authToken = authToken
		}
	}

	return resp, err
}
