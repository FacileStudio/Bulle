package runner

import (
	"context"
	"encoding/json"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

type mailInput struct {
	Account int64 `json:"account"`
	suite.CourrierMessage
}

// mailPolicy shows the owner the email as it would leave: mailbox, threaded
// subject and recipient resolved. When Courrier cannot preview it, the card
// falls back to the model's raw arguments rather than hiding the call.
type mailPolicy struct {
	courrier suite.Courrier
	userID   int64
	defaults suite.Defaults
}

func (p mailPolicy) describe(ctx context.Context, input json.RawMessage) (map[string]any, string, error) {
	var in mailInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, "", err
	}
	var source any = input
	if preview, err := p.courrier.Preview(ctx, p.userID, p.defaults.DefaultAccount(in.Account), in.CourrierMessage); err == nil {
		source = preview
	}
	record, err := asRecord(source)
	if err != nil {
		return nil, "", err
	}
	record["kind"] = "mail"
	return record, "", nil
}

func asRecord(value any) (map[string]any, error) {
	data, ok := value.(json.RawMessage)
	if !ok {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		data = encoded
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return record, nil
}

func (p mailPolicy) denied(context.Context, json.RawMessage) {}

func (g *gate) addPolicy(name string, policy approvalPolicy) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.policies == nil {
		g.policies = map[string]approvalPolicy{}
	}
	g.policies[name] = policy
}
