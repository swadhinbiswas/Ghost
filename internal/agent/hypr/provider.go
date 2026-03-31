package hypr

import (
	"bufio"
	"bytes"
	"cmp"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/fantasy"
	"charm.land/fantasy/object"
	"github.com/charmbracelet/crush/internal/event"
)

var embedded []byte

var Enabled=sync.OnceValue(func()bool{
	b,_:=strconv.ParseBool(
		cmp.Or(
			os.Getenv("HYPER"),
			os.Getenv("HYPERCRUSH"),
			os.Getenv("HYPER_ENABLE"),
			os.Getenv("HYPER_ENABLED"),
		),

	)
	return b 
})
