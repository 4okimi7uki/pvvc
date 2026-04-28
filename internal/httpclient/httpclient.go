package httpclient

import (
	"fmt"
	"net/http"

	"github.com/4okimi7uki/pvvc/internal/gh"
)

var resolvedVersion = gh.ResolvedVersion()
var userAgent = fmt.Sprintf("pvvc/%s (github.com/4okimi7uki/pvvc)", resolvedVersion)

type uaTransport struct {
	base http.RoundTripper
}

func (t *uaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", userAgent)
	return t.base.RoundTrip(req)
}

// New returns an *http.Client that sets User-Agent: pvvc on every request.
func New() *http.Client {
	return &http.Client{
		Transport: &uaTransport{base: http.DefaultTransport},
	}
}
