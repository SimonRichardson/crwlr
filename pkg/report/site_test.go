package report

import "testing"

func TestAggregation_Page(t *testing.T) {
	t.Parallel()

	t.Run("aggregatePages", func(t *testing.T) {
		c := map[string]*Page{}

		c["http://a.com/page1"] = &Page{
			Links: []string{"http://b.com"},
		}
		c["http://a.com/page2"] = &Page{
			Links: []string{"http://b.com", "http://c.com"},
		}

		report, err := aggregatePages(c)
		if err != nil {
			t.Fatal(err)
		}

		if x, ok := report["http://a.com/page1"]; !ok {
			t.Errorf("expected url, found nothing.")
		} else {
			if actual := len(x.Links); actual != 1 {
				t.Errorf("expected: %d, actual: %d", 1, actual)
			}
		}

		if x, ok := report["http://a.com/page2"]; !ok {
			t.Errorf("expected url, found nothing.")
		} else {
			if actual := len(x.Links); actual != 2 {
				t.Errorf("expected: %d, actual: %d", 2, actual)
			}
		}
	})

	t.Run("merge aggregatePages", func(t *testing.T) {
		c := map[string]*Page{}

		c["http://a.com/page1?hello"] = &Page{
			Links: []string{"http://b.com", "http://c.com", "http://d.com"},
		}
		c["http://a.com/page1?world"] = &Page{
			Links: []string{"http://e.com", "http://f.com"},
		}

		report, err := aggregatePages(c)
		if err != nil {
			t.Fatal(err)
		}

		if x, ok := report["http://a.com/page1"]; !ok {
			t.Errorf("expected url, found nothing.")
		} else {
			if actual := len(x.Links); actual != 5 {
				t.Errorf("expected: %d, actual: %d", 5, actual)
			}
		}
	})
}
