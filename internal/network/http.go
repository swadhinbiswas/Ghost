package network

import (
	"fmt"
	"net/http"
)

// GatedTransport is an http.RoundTripper that enforces Allow(host) on every
// outbound request. Use it as the Transport on every http.Client Ghost
// creates, so stealth mode is automatically enforced.
type GatedTransport struct {
	Inner http.RoundTripper
}

// NewGatedTransport returns a GatedTransport wrapping http.DefaultTransport.
func NewGatedTransport() *GatedTransport {
	return &GatedTransport{Inner: http.DefaultTransport}
}

// RoundTrip enforces Allow(host).
func (g *GatedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if !Allow(r.URL.Hostname()) {
		return nil, fmt.Errorf("network: %s blocked by stealth policy", r.URL.Hostname())
	}
	inner := g.Inner
	if inner == nil {
		inner = http.DefaultTransport
	}
	return inner.RoundTrip(r)
}

// GatedClient returns an *http.Client that uses GatedTransport.
func GatedClient() *http.Client {
	return &http.Client{Transport: NewGatedTransport()}
}
