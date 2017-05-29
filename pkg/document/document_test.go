package document

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"golang.org/x/net/html"
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
		doc.Walk(func(root *url.URL, node *html.Node) error {
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
		doc.Walk(Links(func(url *url.URL) error {
			actual = append(actual, url.String())
			return nil
		}))

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

func TestWalkAssets(t *testing.T) {
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
		doc.Walk(Assets(func(url *url.URL) error {
			actual = append(actual, url.String())
			return nil
		}))

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
<link rel="stylesheet" href="/styles.css" />
</head>
<body>
<h1>Heading</h1>
<img src="/image.jpg" />
<img src="/image1.gif" />
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
<link rel="stylesheet" href="/styles.css" />
</head>
<body>
<h1>Heading</h1>
<img src="/image.jpg" />
<img src="#image" />
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
<link rel="stylesheet" href="/styles.css" />
<link rel="stylesheet" href="http://url.com/styles1.css" />
</head>
<body>
<h1>Heading</h1>
<img src="/image.jpg" />
<img src="http://google.com/image1.gif" />
<img src="/image1.gif" />
<img src="http://url.com/image2.gif" />
</body>
`
		urls := []string{
			"http://url.com/styles.css",
			"http://url.com/styles1.css",
			"http://url.com/image.jpg",
			"http://google.com/image1.gif",
			"http://url.com/image1.gif",
			"http://url.com/image2.gif",
		}

		if expected, actual := urls, fn(body); !reflect.DeepEqual(expected, actual) {
			t.Errorf("expected: %v, actual: %v", expected, actual)
		}
	})
}

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
		doc    = NewDocument(u, node, log.NewNopLogger())
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		doc.Walk(func(root *url.URL, node *html.Node) error {
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
		doc    = NewDocument(u, node, log.NewNopLogger())
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		doc.Walk(Links(func(url *url.URL) error {
			actual++
			return nil
		}))
	}
}

func BenchmarkAssets(b *testing.B) {
	u, err := url.Parse("http://url.com")
	if err != nil {
		b.Fatal(err)
	}

	body := `
<!DOCTYPE html>
<html>
<head>
<title>Title</title>
<link rel="stylesheet" href="/styles.css" />
</head>
<body>
<h1>Heading</h1>
<a href="/link">link</a>
<a href="/link1">link1</a>
<a href="/link2">link2</a>
<img src="/image.jpg" />
<img src="/image1.gif" />
</body>
`

	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		b.Fatal(err)
	}

	var (
		actual int
		doc    = NewDocument(u, node, log.NewNopLogger())
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		doc.Walk(Links(func(url *url.URL) error {
			actual++
			return nil
		}))
	}
}
