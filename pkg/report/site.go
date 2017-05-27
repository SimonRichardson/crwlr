package report

import (
	"fmt"
	"io"
	"net/url"
)

// SiteReport creates a report about a page of the crawl
type SiteReport struct {
	pages map[string]*Page
}

// NewSiteReport generates a report from the cache
func NewSiteReport(pages map[string]*Page) *SiteReport {
	return &SiteReport{pages}
}

func (r *SiteReport) Write(w io.Writer) error {
	pages, err := aggregatePages(r.pages)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "URL\tRef Links\t")
	for k, v := range pages {
		fmt.Fprintf(w, "%s\t \t\n", k)

		total := len(v.Links)
		for k, v := range v.Links {
			val := "├──"
			if k == total-1 {
				val = "└──"
			}
			fmt.Fprintf(w, "%s\t%s\t\n", val, v)
		}

		total = len(v.Assets)
		for k, v := range v.Assets {
			val := "├──"
			if k == total-1 {
				val = "└──"
			}
			fmt.Fprintf(w, "%s\t%s\t\n", val, v)
		}
	}

	return nil
}

// Page records the state of a page
type Page struct {
	Links  []string
	Assets []string
}

// Add sums pages together
func (p *Page) Add(o *Page) {
	p.Links = append(p.Links, o.Links...)
	p.Assets = append(p.Assets, o.Assets...)
}

// Aggregate, takes a cache and removes any possible duplication and aggregates
// the values.
func aggregatePages(c map[string]*Page) (map[string]*Page, error) {
	m := map[string]*Page{}

	for k, v := range c {
		// We want to normalize the path (k), to remove parameters and hashes.
		u, err := url.Parse(k)
		if err != nil {
			return m, err
		}

		u.RawQuery = ""
		u.Fragment = ""

		val := u.String()

		if r, ok := m[val]; ok {
			r.Add(v)
		} else {
			r := &Page{}
			r.Add(v)
			m[val] = r
		}
	}

	return m, nil
}
