package pkg

import (
	"sync"
	"testing"
	"time"
)

func TestInitMetrics(t *testing.T) {
	InitMetrics()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for range tick.C {
			Metrics.Download(100)
			Metrics.Upload(200)
		}
	}()

	wg.Wait()
}
