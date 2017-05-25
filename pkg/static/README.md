# Static

The static pkg allows for testing of crawling for both manual and integration
testing.

## API

The `API` type uses the `http.FileServer` to enable serving of static files over
a http request.

The `static.go` file contains the files from the `/templates` folder, that can
be either served locally for development or encoded into the actual `.go` file
so when the cli is built it contains the templates for use.

The `filesys.go` is a middleware that sits in between `API` type and `static.go`
file system. The aim of the middleware is to allow the rendering of template
files so it makes it easier to generate lots of files with ease. It can be
configured to generate 100s and 1000s of template files to fully push the
crawler if required.
