package suite

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/FacileStudio/porte/exchange"
)

// exchangeWaits are the pauses between attempts at an exchange Registre did
// not answer, 7.5 s in all.
var exchangeWaits = []time.Duration{500 * time.Millisecond, time.Second, 2 * time.Second, 4 * time.Second}

// registreRecovery is how long after a series of attempts failed a new
// exchange tries once instead of waiting through another series.
const registreRecovery = 10 * time.Second

// outage makes every exchange that meets an unanswering Registre share one
// series of attempts: the first one retries, the others wait for it and then
// try once, so fifteen apps listed for one run cost one series, not fifteen.
type outage struct {
	mu       sync.Mutex
	wait     chan struct{}
	failedAt time.Time
}

func (o *outage) ride(ctx context.Context, now func() time.Time, try func() (string, error), first error) (string, error) {
	o.mu.Lock()
	if wait := o.wait; wait != nil {
		o.mu.Unlock()
		select {
		case <-ctx.Done():
			return "", first
		case <-wait:
		}
		return try()
	}
	if !o.failedAt.IsZero() && now().Sub(o.failedAt) < registreRecovery {
		o.mu.Unlock()
		return "", first
	}
	wait := make(chan struct{})
	o.wait = wait
	o.mu.Unlock()
	token, err := retryExchange(ctx, try, first)
	o.mu.Lock()
	o.wait = nil
	o.failedAt = time.Time{}
	if transientExchange(err) {
		o.failedAt = now()
	}
	o.mu.Unlock()
	close(wait)
	return token, err
}

func (o *outage) end() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.failedAt = time.Time{}
}

func retryExchange(ctx context.Context, try func() (string, error), err error) (string, error) {
	for _, pause := range exchangeWaits {
		select {
		case <-ctx.Done():
			return "", err
		case <-time.After(pause):
		}
		token, next := try()
		if !transientExchange(next) {
			return token, next
		}
		err = next
	}
	return "", err
}

// transientExchange reports whether Registre failed to answer rather than
// refused: it could not be reached, answered 5xx, 429 or 408, or answered 404
// with no OAuth error, which is Traefik while Registre's container swaps.
func transientExchange(err error) bool {
	if errors.Is(err, exchange.ErrUnavailable) {
		return true
	}
	var refused *exchange.RefusedError
	if !errors.As(err, &refused) {
		return false
	}
	switch refused.Status {
	case http.StatusTooManyRequests, http.StatusRequestTimeout:
		return true
	case http.StatusNotFound:
		return refused.Code == "unknown"
	}
	return refused.Status >= http.StatusInternalServerError
}
