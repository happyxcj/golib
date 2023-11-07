package lib

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// go test -bench=. -benchmem -test.v

func TestFnServeF(t *testing.T) {
	fw := NewFnWorker(2, 10) // 总共最大长度20
	if err := fw.Serve(nil); err == nil {
		t.Fatalf("TestfwServeF Serve nil err, get: %v", err)
	}
	var tmpCount int32
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			fw.Serve(func() {
				atomic.AddInt32(&tmpCount, 1)
				wg.Done()
			})
		}(i)
	}
	wg.Wait()
	if tmpCount != 1000 {
		t.Fatalf("TestFnServeF err, want: %v, get: %v", 1000, tmpCount)
	}
}

func TestFnServeWithTimeout(t *testing.T) {
	fw := NewFnWorker(2, 10) // 总共最大长度20
	if err := fw.ServeWithTimeout(nil, time.Second); err == nil {
		t.Fatalf("TestFnServeWithTimeout Serve nil err, get: %v", err)
	}
	var timeoutCount int32
	var wg sync.WaitGroup
	want := 1000 - 2*10 - 2
	wg.Add(want)
	for i := 0; i < 1000; i++ {
		go func(i int) {
			err := fw.ServeWithTimeout(func() {
				select {}
			}, time.Second)
			if err != nil && err.Error() == "failed to serve the fn, caused by timeout" {
				atomic.AddInt32(&timeoutCount, 1)
				wg.Done()
			}
		}(i)
	}
	wg.Wait()
	if timeoutCount != int32(want) {
		t.Fatalf("TestFnServeWithTimeout err, want: %v, get: %v", want, timeoutCount)
	}
}

func TestFnServeRightfway(t *testing.T) {
	fw := NewFnWorker(2, 10) // 总共最大长度20
	if err := fw.ServeRightAway(nil); err == nil {
		t.Fatalf("TestFnServeRightfway Serve nil err, get: %v", err)
	}
	var timeoutCount int32
	var wg sync.WaitGroup
	want := 1000 - 2*10 - 2
	wg.Add(want)
	for i := 0; i < 1000; i++ {
		go func(i int) {
			err := fw.ServeRightAway(func() {
				select {}
			})
			if err != nil && err.Error() == "slow consumer detected" {
				atomic.AddInt32(&timeoutCount, 1)
				wg.Done()
			}
		}(i)
	}
	wg.Wait()
	if timeoutCount != int32(want) {
		t.Fatalf("TestFnServeRightfway err, want: %v, get: %v", want, timeoutCount)
	}
}

func TestFnGracefulStop(t *testing.T) {
	fw := NewFnWorker(2, 10) // 总共最大长度20

	endCh := make(chan int, 0)
	var wg sync.WaitGroup
	for i := 0; i < 22; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fw.ServeRightAway(func() {
				<-endCh
			})
		}(i)
	}
	wg.Wait()
	go func() {
		time.Sleep(time.Second)
		close(endCh)
	}()
	pendingCount, lostCount := fw.GracefulStop()
	if pendingCount != 20 || lostCount != 0 {
		t.Fatalf("TestFnServeRightfway err, want: [%v_%v], get: [%v_%v]", 20, 0, pendingCount, lostCount)
	}
}

func BenchmarkServeF(b *testing.B) {
	fw := NewFnWorker(50, 200)
	var wg sync.WaitGroup
	var count, wantCount uint64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg.Add(1)
			f := func() {
				atomic.AddUint64(&count, 1)
				wg.Done()
			}
			fw.Serve(f)
			atomic.AddUint64(&wantCount, 1)
		}
	})
	wg.Wait()
	if count != wantCount {
		b.Fatalf("failed to add count, want: %v, get: %v", wantCount, count)
	}
}
