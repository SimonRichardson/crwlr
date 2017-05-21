all: pkg/static/static.go dist/crwlr

pkg/static/static.go:
	esc -o="pkg/static/static.go" -pkg="static" -private templates

dist/crwlr:
	go build -o dist/crwlr github.com/SimonRichardson/crwlr/cmd/crwlr

clean: FORCE
	rm pkg/static/static.go
	rm -rf dist

FORCE:
