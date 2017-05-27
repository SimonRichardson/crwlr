package document

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/go-kit/kit/log"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Document wraps a html document (page) so that we can extract all the relavent
// urls from it.
type Document struct {
	url    *url.URL
	node   *html.Node
	logger log.Logger
}

// NewDocument creates a new Document to use.
func NewDocument(url *url.URL, node *html.Node, logger log.Logger) *Document {
	return &Document{
		url:    url,
		node:   node,
		logger: logger,
	}
}

// Walk walks through the Document node by node.
func (d *Document) Walk(fn Walker) error {
	var f func(*html.Node) error
	f = func(n *html.Node) error {
		if n.Type == html.ElementNode {
			if err := fn(d.url, n); err != nil {
				return err
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := f(c); err != nil {
				return err
			}
		}

		return nil
	}
	return f(d.node)
}

// Walker describes a type that can walk over a documents nodes.
type Walker func(*url.URL, *html.Node) error

// Links walks through all the Documents nodes links.
// Note: it will normalize the documents links to the documents url.
func Links(fn func(*url.URL) error) Walker {
	return func(root *url.URL, node *html.Node) error {
		// We only care about "anchor" links
		if node.DataAtom == atom.A {
			for _, a := range node.Attr {
				// Pluck the "href" from all "a" links.
				if a.Key == "href" {
					if u, ok := normalizeLink(root, a.Val); ok {
						fn(u)
					}
				}
			}
		}

		return nil
	}
}

// Assets walks through all the Documents nodes static assets.
// Note: it will normalize the documents assets urls to the documents url.
func Assets(fn func(*url.URL) error) func(*url.URL, *html.Node) error {
	return func(root *url.URL, node *html.Node) error {
		// We only care about "anchor" links
		switch node.DataAtom {
		case atom.Img:
			for _, a := range node.Attr {
				// Pluck the "src" from all "img" links.
				if a.Key == "src" {
					if u, ok := normalizeLink(root, a.Val); ok {
						fn(u)
					}
				}
			}
		case atom.Link:
			var found bool
			for _, a := range node.Attr {
				if a.Key == "rel" && a.Val == "stylesheet" {
					found = true
					break
				}
			}

			if found {
				for _, a := range node.Attr {
					if a.Key == "href" {
						if u, ok := normalizeLink(root, a.Val); ok {
							fn(u)
						}
					}
				}
			}
		}

		return nil
	}
}

// Compose attempts to compose two walkers together to allow a very basic
// loop fusion.
func Compose(fn1, fn2 Walker) Walker {
	return func(root *url.URL, node *html.Node) error {
		if err := fn1(root, node); err != nil {
			return err
		}
		return fn2(root, node)
	}
}

func normalizeLink(root *url.URL, val string) (*url.URL, bool) {
	// This is a page anchor tag, we don't care about these.
	if strings.HasPrefix(val, "#") {
		return nil, false
	}

	// If it's a relative url to the root, then normalize it.
	n, err := url.Parse(val)
	if err != nil {
		return nil, false
	}

	var norm *url.URL
	if strings.HasPrefix(val, "/") {
		norm = root.ResolveReference(n)
	} else {
		norm = n
	}

	return norm, true

}

// Extract the scheme and host out of the url.
// Note this is part copied from the go stdlib
func host(u *url.URL) string {
	var buf bytes.Buffer
	if u.Scheme != "" {
		buf.WriteString(u.Scheme)
		buf.WriteByte(':')
	}
	if u.Opaque != "" {
		buf.WriteString(u.Opaque)
	} else {
		if u.Scheme != "" || u.Host != "" || u.User != nil {
			buf.WriteString("//")
			if ui := u.User; ui != nil {
				buf.WriteString(ui.String())
				buf.WriteByte('@')
			}
			if h := u.Host; h != "" {
				buf.WriteString(h)
			}
		}
	}
	return buf.String()
}
