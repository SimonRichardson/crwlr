all: dist/crwlr

dist/crwlr:
	go build -o dist/crwlr github.com/SimonRichardson/crwlr/cmd/crwlr

clean: FORCE
	rm -rf dist

FORCE:
