package import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestProbabilisticEarlyExpiry_ShouldRecompute(t *testing.T) {
	memCache := NewInMemoryCache(100 * time.Millisecond)
	per := NewProbabilisticEarlyExpiry(memCache, 1.0)

	// Hard expired entry
	expiredEntry := StampedeEntry{
		Delta:      100 * time.Millisecond,
		Expiration: time.Now().Add(-1 * time.Second),
	}
	if !per.ShouldRecompute(expiredEntry) {
		t.Error("expected ShouldRecompute=true for expired entry")
	}

	// Zero expiration (indefinite TTL)
	zeroExpEntry := StampedeEntry{
		Delta:      100 * time.Millisecond,
		Expiration: time.Time{},
	}
	if per.ShouldRecompute(zeroExpEntry) {
		t.Error("expected ShouldRecompute=false for zero expiration entry")
	}
}

func TestProbabilisticEarlyExpiry_SingleFlightCoalescing(t *testing.T) {
	memCache := NewInMemoryCache(100 * time.Millisecond)
	per := NewProbabilisticEarlyExpiry(memCache, 1.0)

	var fetchCount int32
	fetcher := func() (interface{}, time.Duration, error) {
		atomic.AddInt32(&fetchCount, 1)
		time.Sleep(50 * time.Millisecond)
		return "computed-data", 10 * time.Second, nil
	}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			val, err := per.GetOrFetch("shared-key", fetcher)
			if err != nil {
				t.Errorf("GetOrFetch failed: %v", err)
			}
			if val != "computed-data" {
				t.Errorf("expected computed-data, got %v", val)
			}
		}()
	}

	wg.Wait()

	if atomic.LoadInt32(&fetchCount) != 1 {
		t.Errorf("expected singleflight coalescing to 1 fetch call, got %d", fetchCount)
	}
}

func TestProbabilisticEarlyExpiry_FetcherError(t *testing.T) {
	memCache := NewInMemoryCache(100 * time.Millisecond)
	per := NewProbabilisticEarlyExpiry(memCache, 1.0)

	errFetcher := func() (interface{}, time.Duration, error) {
		return nil, 0, errors.New("origin database down")
	}

	_, err := per.GetOrFetch("err-key", errFetcher)
	if err == nil {
		t.Error("expected error from fetcher")
	}
}
