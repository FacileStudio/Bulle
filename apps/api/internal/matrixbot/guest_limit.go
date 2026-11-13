package matrixbot

import (
	"strconv"
	"sync"
	"time"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
)

const (
	guestBurst  = 5
	guestWindow = 10 * time.Minute
	guestKeys   = 10_000
)

// slowDown answers a colleague who wrote more than guestBurst messages to the
// owner's bot inside guestWindow.
const slowDown = "You have sent several messages in a short time. Try again in a few minutes."

// guestLimiter caps how often one colleague can run the owner's model, so a
// guest cannot spend the owner's money or fill the owner's run slots: a
// bucket of guestBurst messages per owner and sender that refills over
// guestWindow, with at most guestKeys pairs remembered.
type guestLimiter struct {
	once    sync.Once
	limiter *troncmiddleware.Limiter
}

func (l *guestLimiter) allow(owner int64, sender string, now time.Time) bool {
	l.once.Do(func() {
		l.limiter = troncmiddleware.NewLimiter(troncmiddleware.RatelimitConfig{
			Limit: guestBurst / guestWindow.Seconds(), Burst: guestBurst, MaxKeys: guestKeys,
		})
	})
	allowed, _ := l.limiter.AllowAt(strconv.FormatInt(owner, 10)+"\x00"+sender, now)
	return allowed
}
