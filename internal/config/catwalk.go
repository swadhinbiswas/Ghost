package config

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/catwalk/pkg/embedded"
)

type catwalkClient interface {
	GetProviders(context.Context, string) ([]catwalk.Provider, error)
}

type catwalkSync struct {
	cache      cache[[]catwalk.Provider]
	client     catwalkClient
	autoupdate bool
	init       atomic.Bool
}

func (s *catwalkSync) Init(client catwalkClient, path string, autoupdate bool) {
	s.client = client
	s.cache = newCache[[]catwalk.Provider](path)
	s.autoupdate = autoupdate
	s.init.Store(true)
}

func (s *catwalkSync) Get(ctx context.Context) ([]catwalk.Provider, error) {
	if !s.init.Load() {
		panic("called Get before Init")
	}

	cached, etag, cachedErr := s.cache.Get()
	if len(cached) == 0 || cachedErr != nil {
		// if cached file is empty, default to embedded providers
		cached = embedded.GetAll()
	}

	if !s.autoupdate {
		slog.Info("Using cached/embedded Catwalk providers")
		return cached, nil
	}

	slog.Info("Fetching providers from Catwalk")
	result, err := s.client.GetProviders(ctx, etag)
	if errors.Is(err, context.DeadlineExceeded) {
		slog.Warn("Catwalk providers not updated in time")
		return cached, nil
	}
	if errors.Is(err, catwalk.ErrNotModified) {
		slog.Info("Catwalk providers not modified")
		return cached, nil
	}
	if err != nil {
		// On error, fall back to cached (which defaults to embedded if empty).
		return cached, nil
	}
	if len(result) == 0 {
		return cached, errors.New("empty providers list from catwalk")
	}

	return result, s.cache.Store(result)
}
