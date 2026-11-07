package approval

import (
	"context"
	"encoding/json"

	"github.com/FacileStudio/nacelle"
)

func (a *Approvals) allows(ctx context.Context, name string, input json.RawMessage) bool {
	return a.Ask(ctx, nacelle.Call{ID: "c1", Name: name, Input: input}).Deny == ""
}

func (a *Approvals) accepts(ctx context.Context, name string, input json.RawMessage) bool {
	return a.Accept(ctx, nacelle.Call{ID: "c1", Name: name, Input: input}).Deny == ""
}
