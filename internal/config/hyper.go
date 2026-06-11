package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	xetag "github.com/charmbracelet/x/etag"
)

type hyperClient interface {
	Get(context.Context, string) (catwalk.Provider, error)
}

var _ hyperClient = realHyperClient{}

type realHyperClient struct {
	baseURL string
}

// Get implements hyperClient.
func (r realHyperClient) Get(ctx context.Context, etag string) (catwalk.Provider, error) {
	var result catwalk.Provider
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		r.baseURL+"/api/v1/provider",
		nil,
	)
	if err != nil {
		return result, fmt.Errorf("could not create request: %w", err)
	}
	xetag.Request(req, etag)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == http.StatusNotModified {
		return result, catwalk.ErrNotModified
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
