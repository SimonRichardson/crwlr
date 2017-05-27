package report

import (
	"fmt"
	"io"
	"net/url"
	"time"
)

// MetricReport creates a report about the mertics of a crawl
type MetricReport struct {
	duration time.Duration
	rows     map[string]*Row
}

// NewMetricReport generates a report from the cache including a duration
func NewMetricReport(duration time.Duration, rows map[string]*Row) *MetricReport {
	return &MetricReport{duration, rows}
}

func (r *MetricReport) Write(w io.Writer) error {
	rows, err := aggregateRows(r.rows)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, " URL\t Avg Duration (ms)\t Requested\t Received\t Filtered\t Errorred\t")
	for k, v := range rows {
		fmt.Fprintf(w, " %s\t %d\t %d\t %d\t %d\t %d\t\n",
			k,
			v.Duration.Nanoseconds()/1e6,
			v.Requested,
			v.Received,
			v.Filtered,
			v.Errorred,
		)
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, " Totals\t Duration (ms)\t")
	fmt.Fprintf(w, " \t %d\t\n", r.duration.Nanoseconds()/1e6)

	return nil
}

// Row is used for metric reporting
type Row struct {
	Requested, Received     int
	Filtered, Errorred      int
	TotalDuration, Duration time.Duration
}

// Add sums metrics to the column data
func (c *Row) Add(m *Row) {
	c.Requested += m.Requested
	c.Received += m.Received
	c.Filtered += m.Filtered
	c.Errorred += m.Errorred

	c.TotalDuration += m.Duration
	c.Duration = c.TotalDuration
	// Prevent division by zero
	if c.Received > 0 {
		c.Duration = c.TotalDuration / time.Duration(c.Received)
	}
}

// Aggregate, takes a cache and removes any possible duplication and aggregates
// the values.
func aggregateRows(c map[string]*Row) (map[string]*Row, error) {
	m := map[string]*Row{}

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
			r := &Row{}
			r.Add(v)
			m[val] = r
		}
	}

	return m, nil
}
