package static

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// API serves the static API.
type API struct {
	fs     http.FileSystem
	logger log.Logger
}

// NewAPI returns a usable API for static templating content.
func NewAPI(local bool, logger log.Logger) *API {
	return &API{
		fs:     TemplateFS(local, logger),
		logger: logger,
	}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	level.Info(a.logger).Log("url", r.URL.String())

	// Make sure the crawler follows 301 redirects
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/index", http.StatusMovedPermanently)
		return
	}

	http.FileServer(a.fs).ServeHTTP(w, r)
}
