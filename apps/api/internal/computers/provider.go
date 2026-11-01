package computers

import (
	"context"

	"github.com/FacileStudio/Bulle/core/sandbox"
)

type Provider interface {
	Ensure(ctx context.Context, userID string) (*sandbox.Target, error)
	Touch(userID string)
}
