package crawler

import (
	"fmt"
	"io"
	"net/url"
	"time"
)

type Report struct {
	duration time.Duration
	cache    *Cache
}

// NewReport generates a report from the cache including a duration
func NewReport(duration time.Duration, cache *Cache) *Report {
	return &Report{duration, cache}
}

func (r *Report) Write(w io.Writer) error {
	rows, err := aggregate(r.cache)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "URL\tDuration (ms)\tRequested\tReceived\tFiltered\tErrorred\t")
	for k, v := range rows {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t\n",
			k,
			v.Duration.Nanoseconds()/1e6,
			v.Requested,
			v.Received,
			v.Filtered,
			v.Errorred,
		)
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Totals\tDuration (ms)\t")
	fmt.Fprintf(w, "\t%d\t\n", r.duration.Nanoseconds()/1e6)

	return nil
}

type row struct {
	Requested, Received     int
	Filtered, Errorred      int
	TotalDuration, Duration time.Duration
}

// Add writes metrics to the column data
func (c *row) Add(m *Metric) {
	c.Requested += int(m.Requested.Time())
	c.Received += int(m.Received.Time())
	c.Filtered += int(m.Filtered.Time())
	c.Errorred += int(m.Errorred.Time())

	c.TotalDuration += m.Duration
	c.Duration = c.TotalDuration / time.Duration(c.Received)
}

// Aggregate, takes a cache and removes any possible duplication and aggregates
// the values.
func aggregate(c *Cache) (map[string]*row, error) {
	m := map[string]*row{}

	for k, v := range c.metrics {
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
			r := &row{}
			r.Add(v)
			m[val] = r
		}
	}

	return m, nil
}
