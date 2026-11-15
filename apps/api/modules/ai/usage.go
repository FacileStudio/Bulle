package ai

import (
	"context"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/core/jev"
	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/google/jsonschema-go/jsonschema"
)

// record writes the call to the Ledger, when there is one.
func (c Config) record(ctx context.Context, userID int64, source, caller string, call aiusage.Call) {
	if c.Ledger != nil {
		c.Ledger.Record(ctx, userID, source, caller, call)
	}
}

// overBudget answers the Budget refusal for a call on the server's key, nil
// when there is no Budget.
func (c Config) overBudget(ctx context.Context, userID int64) error {
	if c.Budget == nil {
		return nil
	}
	return c.Budget.Check(ctx, userID)
}

func decided(response jev.Response) aiusage.Call {
	usage := nacelle.Usage{InputTokens: int64(response.Usage.InputTokens), OutputTokens: int64(response.Usage.OutputTokens), Cost: response.Usage.Cost}
	return aiusage.Call{Model: response.Model, Usage: usage}
}

// readComplete decodes and validates the completion request, and answers the
// error itself when it is refused.
func readComplete(w http.ResponseWriter, r *http.Request) (CompleteRequest, *jsonschema.Resolved, bool) {
	var body CompleteRequest
	err := httpjson.DecodeJSONLimit(w, r, &body, maxCompleteBody)
	var schema *jsonschema.Resolved
	if err == nil {
		schema, err = validateComplete(&body)
	}
	if err != nil {
		httpjson.WriteError(w, err)
	}
	return body, schema, err == nil
}
