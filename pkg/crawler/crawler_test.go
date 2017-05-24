package crawler

import "testing"
import "sync"

func TestGauge(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		gauge := NewGauge()

		for i := 0; i < 100; i++ {
			gauge.Increment()
		}

		for i := 0; i < 100; i++ {
			gauge.Decrement()
		}

		if val := gauge.Value(); val != 0 {
			t.Errorf("expected: 100, actual: %d", val)
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		gauge := NewGauge()

		wg := sync.WaitGroup{}
		wg.Add(200)
		for i := 0; i < 100; i++ {
			go func() {
				defer wg.Done()
				gauge.Increment()
			}()
			go func() {
				defer wg.Done()
				gauge.Decrement()
			}()
		}

		wg.Wait()

		if val := gauge.Value(); val != 0 {
			t.Errorf("expected: 100, actual: %d", val)
		}
	})
}
