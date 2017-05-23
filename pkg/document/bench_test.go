package document_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/SimonRichardson/crwlr/pkg/document"
	"github.com/go-kit/kit/log"
	"golang.org/x/net/html"
)

func BenchmarkWalk(b *testing.B) {
	u, err := url.Parse("http://url.com")
	if err != nil {
		b.Fatal(err)
	}

	body := `
<!DOCTYPE html>
<html>
<head>
<title>Title</title>
</head>
<body>
<h1>Heading</h1>
<a href="/link">link</a>
<a href="/link1">link1</a>
<a href="/link2">link2</a>
</body>
`

	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		b.Fatal(err)
	}

	var (
		actual int
		doc    = document.NewDocument(u, node, log.NewNopLogger())
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		doc.Walk(func(node *html.Node, err error) error {
			actual++
			return nil
		})
	}
}

func BenchmarkLinks(b *testing.B) {
	u, err := url.Parse("http://url.com")
	if err != nil {
		b.Fatal(err)
	}

	body := `
<!DOCTYPE html>
<html>
<head>
<title>Title</title>
</head>
<body>
<h1>Heading</h1>
<a href="/link">link</a>
<a href="/link1">link1</a>
<a href="/link2">link2</a>
</body>
`

	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		b.Fatal(err)
	}

	var (
		actual int
		doc    = document.NewDocument(u, node, log.NewNopLogger())
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		doc.WalkLinks(func(url *url.URL, err error) error {
			actual++
			return nil
		})
	}
}
