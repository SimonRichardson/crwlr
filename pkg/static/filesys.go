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
)

// FileSys returns a http.FileSystem for embedded assets. If useLocal is true,
// the filesystem's content are used instead.
func FileSys(useLocal bool) http.FileSystem {
	if useLocal {
		return localFS{map[string]*file{
			"/index": &file{
				name:  "index",
				local: "templates/index.tmpl",
				data: map[string]string{
					"Heading": "Index",
					"Link":    "/page",
				},
			},
			"/page": &file{
				name:  "page",
				local: "templates/index.tmpl",
				data: map[string]string{
					"Heading": "Page",
					"Link":    "/index",
				},
			},
		}}
	}
	panic("TODO")
}

type file struct {
	name    string
	local   string
	data    interface{}
	size    int64
	modtime int64
	isDir   bool

	once  sync.Once
	bytes []byte
}

type localFS struct {
	data map[string]*file
}

func (fs localFS) Open(name string) (http.File, error) {
	f, ok := fs.data[path.Clean(name)]
	if !ok {
		return nil, os.ErrNotExist
	}

	var err error
	f.once.Do(func() {
		var file http.File
		if file, err = os.Open(f.local); err != nil {
			return
		}

		var b []byte
		if b, err = ioutil.ReadAll(file); err != nil {
			return
		}

		var tmpl *template.Template
		if tmpl, err = template.New(f.name).Parse(string(b)); err != nil {
			return
		}

		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, f.data); err != nil {
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
func (f *httpFile) IsDir() bool                              { return f.isDir }
func (f *httpFile) Sys() interface{}                         { return f }
