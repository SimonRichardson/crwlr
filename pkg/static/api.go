package static

import (
	"net/http"

	"github.com/go-kit/kit/log"
)

// API serves the static API.
type API struct {
	fs     http.FileSystem
	logger log.Logger
}

// NewAPI returns a usable API for static templating content.
func NewAPI(local bool, logger log.Logger) *API {
	return &API{
		fs:     FileSys(local),
		logger: logger,
	}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.FileServer(a.fs).ServeHTTP(w, r)
}
