package static

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// TemplateFS returns a http.FileSystem for embedded assets and templates. This
// filesystem wraps the `github.com/mjibson/esc` API to provide a way to quickly
// load templates from the local file system and also use embedded files within
// the binary once distributed.
// If useLocal is true, the filesystem's content are used instead.
func TemplateFS(useLocal bool, logger log.Logger) http.FileSystem {
	var fs http.FileSystem
	if useLocal {
		fs = _escLocal
	} else {
		fs = _escStatic
	}

	return templateFS{
		fs: fs,
		data: map[string]*file{
			"/index": &file{
				name:     "index",
				local:    "/templates/index.tmpl",
				template: true,
				data: page{
					Title:   "Index",
					Heading: "Index",
					Links: []link{
						link{"/index", "Circular Reference", "Self"},
						link{"/page", "Page", "Page"},
						link{"/bad", "Bad Link", "Bad"},
						link{"http://google.com", "External Link", "External"},
						link{"#anchor", "Anchor", "Anchor"},
					},
				},
			},
			"/page": &file{
				name:     "page",
				local:    "/templates/index.tmpl",
				template: true,
				data: page{
					Title:   "Page",
					Heading: "Page",
					Links: []link{
						link{"/page", "Circular Reference", "Self"},
						link{"/index", "Index", "Index"},
					},
				},
			},
		},
		logger: logger,
	}
}

type file struct {
	name     string
	local    string
	data     page
	size     int64
	modtime  int64
	template bool

	once  sync.Once
	bytes []byte
}

type page struct {
	Title, Heading string
	Links          []link
}

type link struct {
	HREF, ATTR string
	Text       string
}

type templateFS struct {
	fs     http.FileSystem
	data   map[string]*file
	logger log.Logger
}

func (t templateFS) Open(name string) (http.File, error) {
	// Check to see if we have any cached content for the corresponding name.
	// If we do, then check to see if there are any bytes available to send
	// back.
	f, ok := t.data[path.Clean(name)]
	if ok {
		if len(f.bytes) > 0 {
			return &httpFile{
				Reader: bytes.NewReader(f.bytes),
				file:   f,
			}, nil
		}

		// Because we're doing re-writing of urls here, we need to change the
		// name path.
		name = f.local
	}

	// Open the file from the underlying FileSystem
	file, err := t.fs.Open(name)
	if err != nil {
		return nil, err
	}

	// Now check to see if the file should be processed or not
	if !ok || !f.template {
		return file, nil
	}

	// If the file is a template, execute it. The template is only executed once
	// and cached directly on the file for reading later.
	f.once.Do(func() {
		var b []byte
		if b, err = ioutil.ReadAll(file); err != nil {
			return
		}

		var tmpl *template.Template
		if tmpl, err = template.New(f.name).Parse(string(b)); err != nil {
			level.Error(t.logger).Log("err", err)
			return
		}

		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, f.data); err != nil {
			level.Error(t.logger).Log("err", err)
			return
		}

		f.bytes = buf.Bytes()
		f.size = int64(len(f.bytes))
		f.modtime = time.Now().Unix()
	})
	if err != nil {
		return nil, err
	}

	return &httpFile{
		Reader: bytes.NewReader(f.bytes),
		file:   f,
	}, nil
}

type httpFile struct {
	*bytes.Reader
	*file
}

func (f *httpFile) Close() error                             { return nil }
func (f *httpFile) Readdir(count int) ([]os.FileInfo, error) { return nil, nil }
func (f *httpFile) Stat() (os.FileInfo, error)               { return f, nil }
func (f *httpFile) Name() string                             { return f.name }
func (f *httpFile) Size() int64                              { return f.size }
func (f *httpFile) Mode() os.FileMode                        { return 0 }
func (f *httpFile) ModTime() time.Time                       { return time.Unix(f.modtime, 0) }
func (f *httpFile) IsDir() bool                              { return false }
func (f *httpFile) Sys() interface{}                         { return f }
