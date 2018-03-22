package utils

import (
	"sort"

	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

type Throttle interface {
	Limit() rate.Limit
	Wait(ctx context.Context) error
	Allow() bool
}

type throttle struct {
	limiter *rate.Limiter
}

func NewThrottle(limit rate.Limit, burst int) *throttle {
	return &throttle{limiter: rate.NewLimiter(limit, burst)}
}

func (t *throttle) Wait(ctx context.Context) error {
	return t.limiter.Wait(ctx)
}

func (t *throttle) Allow() bool {
	return t.limiter.Allow()
}

func (t *throttle) Limit() rate.Limit {
	return t.limiter.Limit()
}

type multiThrottle struct {
	throttles []Throttle
}

func NewMultiThrottle(throttles ...Throttle) Throttle {
	byLimit := func(i, j int) bool {
		return throttles[i].Limit() < throttles[j].Limit()
	}

	sort.Slice(throttles, byLimit)
	return &multiThrottle{throttles: throttles}
}

func (mt *multiThrottle) Wait(ctx context.Context) error {
	for i := range mt.throttles {
		if err := mt.throttles[i].Wait(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (mt *multiThrottle) Allow() bool {
	for i := range mt.throttles {
		if mt.throttles[i].Allow() == false {
			return false
		}
	}

	return true
}

func (mt *multiThrottle) Limit() rate.Limit {
	if len(mt.throttles) == 0 {
		return 0
	}

	return mt.throttles[0].Limit()
}
