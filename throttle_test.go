package throttle

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestThrottleAllow(t *testing.T) {
	testCases := []struct {
		name   string
		events int64
		burst  int
	}{
		{"1/sec", 1, 10},
		{"1000/sec", 1000, 10},
		{"10000/sec", 10000, 100},
		{"100000/sec", 100000, 1000},
	}

	for _, testCase := range testCases {
		fmt.Println(testCase.name)
		limit := float64(testCase.events) / float64(time.Second.Seconds())
		throttle := New(rate.Limit(limit), testCase.burst)

		var wg sync.WaitGroup
		wg.Add(2)

		getAllBurst := func() <-chan struct{} {
			signal := make(chan struct{})
			go func() {
				defer wg.Done()
				defer close(signal)

				for i := 0; i < testCase.burst; i++ {
					if r := throttle.Allow(); r != true {
						t.Error(testCase.name, "Cannot get next token but should be available")
					}
				}
			}()

			return signal
		}

		checkAfterTimePeriod := func(signal <-chan struct{}) {
			defer wg.Done()
			wait := time.Second.Nanoseconds() / int64(testCase.events)
			select {
			case <-signal:
				<-time.After(time.Duration(wait))
				if r := throttle.Allow(); r != true {
					t.Error("Cannot get next token after waiting but should be available")
				}
			}
		}

		checkAfterTimePeriod(getAllBurst())

		wg.Wait()
	}
}

func TestThrottleWaitContextCancellation(t *testing.T) {
	events := 1
	burst := 1
	limit := float64(events) / float64(time.Second.Seconds())
	throttle := New(rate.Limit(limit), burst)

	ctx, cancel := context.WithCancel(context.Background())

	time.AfterFunc(time.Millisecond, func() { cancel() })

	throttle.Wait(ctx)
	err := throttle.Wait(ctx)

	if err != context.Canceled {
		t.Error("Wait should return that it was cancelled")
	}
}
