package report

import (
	"testing"

	"time"
)

func TestRow(t *testing.T) {
	t.Parallel()

	t.Run("add", func(t *testing.T) {
		r := &Row{}

		r.Add(&Row{
			Received: 1,
			Duration: time.Second,
		})
		r.Add(&Row{
			Received: 1,
			Errorred: 1,
			Duration: time.Second * 3,
		})

		if r.Received != 2 {
			t.Errorf("expected: %d, actual: %d", 2, r.Received)
		}
		if r.Errorred != 1 {
			t.Errorf("expected: %d, actual: %d", 1, r.Errorred)
		}
		if r.Duration != time.Second*2 {
			t.Errorf("expected: %d, actual: %d", time.Second*2, r.Duration)
		}
	})
}

func TestAggregation_Row(t *testing.T) {
	t.Parallel()

	t.Run("aggregateRows", func(t *testing.T) {
		c := map[string]*Row{}

		c["http://a.com/page1"] = &Row{
			Received: 1,
		}
		c["http://a.com/page2"] = &Row{
			Received: 2,
		}

		report, err := aggregateRows(c)
		if err != nil {
			t.Fatal(err)
		}

		if x, ok := report["http://a.com/page1"]; !ok {
			t.Errorf("expected url, found nothing.")
		} else {
			if x.Received != 1 {
				t.Errorf("expected: %d, actual: %d", 1, x.Received)
			}
		}

		if x, ok := report["http://a.com/page2"]; !ok {
			t.Errorf("expected url, found nothing.")
		} else {
			if x.Received != 2 {
				t.Errorf("expected: %d, actual: %d", 2, x.Received)
			}
		}
	})

	t.Run("merge aggregateRows", func(t *testing.T) {
		c := map[string]*Row{}

		c["http://a.com/page1?hello"] = &Row{
			Received: 3,
		}
		c["http://a.com/page1?world"] = &Row{
			Received: 2,
		}

		report, err := aggregateRows(c)
		if err != nil {
			t.Fatal(err)
		}

		if x, ok := report["http://a.com/page1"]; !ok {
			t.Errorf("expected url, found nothing.")
		} else {
			if x.Received != 5 {
				t.Errorf("expected: %d, actual: %d", 5, x.Received)
			}
		}
	})

	t.Run("merge aggregateRows duration", func(t *testing.T) {
		c := map[string]*Row{}

		c["http://a.com/page1?hello"] = &Row{
			Received: 3,
			Duration: time.Second,
		}
		c["http://a.com/page1?world"] = &Row{
			Received: 2,
			Duration: time.Minute,
		}

		report, err := aggregateRows(c)
		if err != nil {
			t.Fatal(err)
		}

		if x, ok := report["http://a.com/page1"]; !ok {
			t.Errorf("expected url, found nothing.")
		} else {
			if expected := (time.Second * 12) + (time.Millisecond * 200); x.Duration != expected {
				t.Errorf("expected: %d, actual: %d", expected, x.Duration)
			}
		}
	})
}
