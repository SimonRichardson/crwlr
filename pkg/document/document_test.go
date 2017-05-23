package document

import (
	"net/url"
	"testing"

	"golang.org/x/net/html"

	"strings"

	"reflect"

	"github.com/go-kit/kit/log"
)

func TestWalk(t *testing.T) {
	t.Parallel()

	fn := func(body string) int {
		u, err := url.Parse("http://url.com")
		if err != nil {
			t.Fatal(err)
		}

		node, err := html.Parse(strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}

		var (
			actual int
			doc    = NewDocument(u, node, log.NewNopLogger())
		)
		doc.Walk(func(node *html.Node, err error) error {
			actual++
			return nil
		})

		return actual
	}

	t.Run("basic", func(t *testing.T) {
		body := `
<!DOCTYPE html>
<html>
<head>
<title>Title</title>
</head>
<body>
<h1>Heading</h1>
</body>
`
		if expected, actual := 5, fn(body); expected != actual {
			t.Errorf("expected: %d, actual: %d", expected, actual)
		}
	})

	t.Run("links", func(t *testing.T) {
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
		if expected, actual := 8, fn(body); expected != actual {
			t.Errorf("expected: %d, actual: %d", expected, actual)
		}
	})
}

func TestWalkLinks(t *testing.T) {
	t.Parallel()

	fn := func(body string) []string {
		u, err := url.Parse("http://url.com")
		if err != nil {
			t.Fatal(err)
		}

		node, err := html.Parse(strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}

		var (
			actual []string
			doc    = NewDocument(u, node, log.NewNopLogger())
		)
		doc.WalkLinks(func(url *url.URL, err error) error {
			actual = append(actual, url.String())
			return nil
		})

		return actual
	}

	t.Run("basic", func(t *testing.T) {
		body := `
<!DOCTYPE html>
<html>
<head>
<title>Title</title>
</head>
<body>
<h1>Heading</h1>
</body>
`
		if expected, actual := 0, len(fn(body)); expected != actual {
			t.Errorf("expected: %d, actual: %d", expected, actual)
		}
	})

	t.Run("links", func(t *testing.T) {
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
		if expected, actual := 3, len(fn(body)); expected != actual {
			t.Errorf("expected: %d, actual: %d", expected, actual)
		}
	})

	t.Run("anchors", func(t *testing.T) {
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
<a href="#anchor">anchor</a>
</body>
`
		if expected, actual := 2, len(fn(body)); expected != actual {
			t.Errorf("expected: %d, actual: %d", expected, actual)
		}
	})

	t.Run("urls", func(t *testing.T) {
		body := `
<!DOCTYPE html>
<html>
<head>
<title>Title</title>
</head>
<body>
<h1>Heading</h1>
<a href="/link">link</a>
<a href="http://url.com/link1">link1</a>
<a href="/link2">link2</a>
<a href="http://google.com/link3">link3</a>
</body>
`
		urls := []string{
			"http://url.com/link",
			"http://url.com/link1",
			"http://url.com/link2",
			"http://google.com/link3",
		}

		if expected, actual := urls, fn(body); !reflect.DeepEqual(expected, actual) {
			t.Errorf("expected: %v, actual: %v", expected, actual)
		}
	})
}
