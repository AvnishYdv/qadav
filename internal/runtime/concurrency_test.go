package runtime

import (
	"sync"
	"testing"
)

func TestConcurrency_ThreadSafety(t *testing.T) {
	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Concurrently compute hash of embedded runtime sources
			h := Hash()
			if len(h) != 16 {
				t.Errorf("goroutine %d: invalid hash length %d", id, len(h))
			}
			// Concurrently retrieve runtime sources
			_, content, err := Sources()
			if err != nil || len(content) == 0 {
				t.Errorf("goroutine %d: failed to read sources: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
}

