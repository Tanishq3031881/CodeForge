// Package ws proxies authenticated WebSocket connections to the Yjs sidecar.
//
// The sidecar trusts every connection it receives, so it must only be
// reachable through this proxy: it binds to localhost in dev and an internal
// Docker network in prod. The Go side verifies the JWT and room access before
// any bytes reach the sidecar.
package ws

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	target *url.URL
}

func NewProxy(rawURL string) (*Proxy, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &Proxy{target: u}, nil
}

// Forward proxies the request to the sidecar, rewriting the path to the
// canonical doc name (the file ID). The query string — which carries the
// JWT — is stripped so the token never reaches the sidecar or its logs.
// httputil.ReverseProxy handles the WebSocket upgrade natively.
func (p *Proxy) Forward(w http.ResponseWriter, r *http.Request, docName string) {
	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.Out.URL.Scheme = p.target.Scheme
			pr.Out.URL.Host = p.target.Host
			pr.Out.URL.Path = "/" + docName
			pr.Out.URL.RawQuery = ""
		},
	}
	rp.ServeHTTP(w, r)
}
