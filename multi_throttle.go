package throttle

import (
	"sort"

	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

type multiThrottle struct {
	throttles []Throttle
}

func NewMulti(throttles ...Throttle) Throttle {
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
