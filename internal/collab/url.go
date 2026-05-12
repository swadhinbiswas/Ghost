package collab

import (
	"net/url"
	"strings"
)

const (
	// DefaultListenAddr is the local bind address for the collaboration server.
	DefaultListenAddr = ":8787"

	// DefaultShareURL is the default base URL used when generating share URLs.
	DefaultShareURL = "ws://localhost:8787"
)

// ShareURL builds a WebSocket URL for the given collaboration room.
func ShareURL(baseURL, roomID string) string {
	if baseURL == "" {
		baseURL = DefaultShareURL
	}
	if !strings.Contains(baseURL, "://") {
		baseURL = "ws://" + baseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		u = &url.URL{Scheme: "ws", Host: "localhost:8787", Path: "/"}
	}
	if u.Path == "" {
		u.Path = "/"
	}

	q := u.Query()
	q.Set("room", roomID)
	u.RawQuery = q.Encode()
	return u.String()
}
