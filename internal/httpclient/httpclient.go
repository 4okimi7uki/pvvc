package httpclient

import "net/http"

const userAgent = "pvvc"

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
