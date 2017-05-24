package document

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
func (d *Document) Walk(fn func(*html.Node) error) error {
	var f func(*html.Node) error
	f = func(n *html.Node) error {
		if n.Type == html.ElementNode {
			if err := fn(n); err != nil {
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

// WalkLinks walks through all the Documents nodes links.
// Note: it will normalize the documents links to the documents url.
func (d *Document) WalkLinks(fn func(*url.URL) error) error {
	return d.Walk(func(node *html.Node) error {
		// We only care about "anchor" links
		if node.DataAtom == atom.A {
			for _, a := range node.Attr {
				// Pluck the "href" from all "a" links.
				if a.Key == "href" {
					// This is a page anchor tag, we don't care about these.
					if strings.HasPrefix(a.Val, "#") {
						break
					}

					// If it's a relative url to the root, then normalize it.
					n, err := url.Parse(a.Val)
					if err != nil {
						level.Debug(d.logger).Log("url", a.Val, "err", err)
						break
					}

					var norm *url.URL
					if strings.HasPrefix(a.Val, "/") {
						norm = d.url.ResolveReference(n)
					} else {
						norm = n
					}

					if err := fn(norm); err != nil {
						level.Debug(d.logger).Log("url", norm, "err", err)
						break
					}
				}
			}
		}

		return nil
	})
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
