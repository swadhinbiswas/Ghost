package hyper

import (
	"cmp"
	_ "embed"
	"errors"
	"net/http"
	"os"
	"strconv"
	"sync"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/fantasy"
)

var embedded []byte

var Enabled = sync.OnceValue(func() bool {
	b, _ := strconv.ParseBool(
		cmp.Or(
			os.Getenv("HYPER"),
			os.Getenv("HYPERGHOST"),
			os.Getenv("HYPER_ENABLE"),
			os.Getenv("HYPER_ENABLED"),
		),
	)
	return b
})

func BaseURL() string {
	return cmp.Or(os.Getenv("HYPER_URL"), "https://api.hyper.land")
}

func Embedded() catwalk.Provider {
	return catwalk.Provider{}
}

const Name = "hyper"

var ErrNoCredits = errors.New("no credits")

type Option func(fantasy.Provider) error

func WithAPIKey(k string) Option {
	return func(p fantasy.Provider) error { return nil }
}

func WithHTTPClient(c *http.Client) Option {
	return func(p fantasy.Provider) error { return nil }
}

func New(opts ...Option) (fantasy.Provider, error) {
	return nil, nil // or basically empty stub since we only want it to compile for now
}
