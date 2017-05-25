package crawler

import (
	"testing"

	"time"

	"github.com/go-kit/kit/log"
)

func TestRow(t *testing.T) {
	t.Parallel()

	t.Run("add", func(t *testing.T) {
		r := row{}

		{
			m := NewMetric()
			m.Received.Increment()
			m.Duration = time.Second
			r.Add(m)
		}
		{
			m := NewMetric()
			m.Received.Increment()
			m.Errorred.Increment()
			m.Duration = time.Second * 3
			r.Add(m)
		}

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

func TestAggregation(t *testing.T) {
	t.Parallel()

	t.Run("aggregate", func(t *testing.T) {
		c := NewCache(log.NewNopLogger())

		{
			m := NewMetric()
			m.Received.Increment()
			c.Set("http://a.com/page1", m)
		}
		{
			m := NewMetric()
			m.Received.Increment()
			m.Received.Increment()
			c.Set("http://a.com/page2", m)
		}

		report, err := aggregate(c)
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

	t.Run("merge aggregate", func(t *testing.T) {
		c := NewCache(log.NewNopLogger())

		{
			m := NewMetric()
			m.Received.Increment()
			m.Received.Increment()
			m.Received.Increment()
			c.Set("http://a.com/page1?hello", m)
		}
		{
			m := NewMetric()
			m.Received.Increment()
			m.Received.Increment()
			c.Set("http://a.com/page1?world", m)
		}

		report, err := aggregate(c)
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

	t.Run("merge aggregate duration", func(t *testing.T) {
		c := NewCache(log.NewNopLogger())

		{
			m := NewMetric()
			m.Received.Increment()
			m.Received.Increment()
			m.Received.Increment()
			m.Duration = time.Second
			c.Set("http://a.com/page1?hello", m)
		}
		{
			m := NewMetric()
			m.Received.Increment()
			m.Received.Increment()
			m.Duration = time.Minute
			c.Set("http://a.com/page1?world", m)
		}

		report, err := aggregate(c)
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
