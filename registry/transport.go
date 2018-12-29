package registry

import (
	"fmt"
	"net/http"
	"time"
)

// AuthTokenTransport defines the data structure for custom http.Request options.
type AuthTokenTransport struct {
	Transport http.RoundTripper
	authToken string
	expiredAt time.Time
}

// RoundTrip defines the round tripper for the error transport.
func (t *AuthTokenTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if t.authToken != "" && t.isFresh() && request.Header.Get("Authorization") == "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.authToken))
	}

	resp, err := t.Transport.RoundTrip(request)
	if err == nil {
		authToken := resp.Header.Get("request-token")
		if authToken != "" {
			t.authToken = authToken
			t.expiredAt = time.Now().Add(4 * time.Minute)
		}
	}

	return resp, err
}

func (t *AuthTokenTransport) isFresh() bool {
	return time.Now().Before(t.expiredAt)
}
