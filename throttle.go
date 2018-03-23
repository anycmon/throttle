package throttle

import (
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

func New(limit rate.Limit, burst int) *throttle {
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
