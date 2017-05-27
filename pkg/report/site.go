package report

import (
	"fmt"
	"io"
	"math"
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

	fmt.Fprintln(w, " URL\t Ref Links\t Ref Assets\t")
	for k, v := range pages {
		fmt.Fprintf(w, " %s\t \t \t\n", k)

		var (
			linkTotal  = len(v.Links)
			assetTotal = len(v.Assets)
			max        = int(math.Max(float64(linkTotal), float64(assetTotal)))
			rows       = make([]*row, max)
		)

		for i := 0; i < max; i++ {
			r := &row{}
			if i < linkTotal {
				r.Link = v.Links[i]
			}
			if i < assetTotal {
				r.Asset = v.Assets[i]
			}
			rows[i] = r
		}

		for _, v := range rows {
			fmt.Fprintf(w, " \t %s\t %s\t\n", v.Link, v.Asset)
		}
	}

	return nil
}

type row struct {
	Link, Asset string
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
