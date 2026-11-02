package suite

import "context"

const sablierRunningPath = "/api/time-entries/running"

type Sablier struct{ gw *Gateway }

func NewSablier(gw *Gateway) Sablier { return Sablier{gw: gw} }

func (s Sablier) Running(ctx context.Context, userID int64) ([]byte, error) {
	return s.gw.Get(ctx, userID, AppSablier, sablierRunningPath, nil)
}
