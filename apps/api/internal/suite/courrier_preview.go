package suite

import "context"

// MailPreview is the email exactly as courrier_send would send it: threading
// already applied, sender resolved to an address.
type MailPreview struct {
	Kind     string   `json:"kind"`
	Account  int64    `json:"account"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	Cc       []string `json:"cc,omitempty"`
	Bcc      []string `json:"bcc,omitempty"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	ReplyTo  int64    `json:"reply_to,omitempty"`
	Threaded bool     `json:"threaded"`
}

// Preview resolves a message the way Send would, without sending it, so the
// approval card shows the real recipient, subject and mailbox.
func (c Courrier) Preview(ctx context.Context, userID, account int64, msg CourrierMessage) (MailPreview, error) {
	id, body, err := c.compose(ctx, userID, account, msg)
	if err != nil {
		return MailPreview{}, err
	}
	preview := MailPreview{
		Kind: "mail", Account: id, To: body.To, Cc: body.Cc, Bcc: body.Bcc,
		Subject: body.Subject, Body: body.Body, ReplyTo: msg.ReplyTo, Threaded: body.InReplyTo != "",
	}
	if accounts, err := c.Accounts(ctx, userID); err == nil {
		for _, acc := range accounts {
			if acc.ID == id {
				preview.From = acc.Email
			}
		}
	}
	return preview, nil
}

// DefaultAccount is the mailbox a tool call with no account uses.
func (d Defaults) DefaultAccount(requested int64) int64 { return d.account(requested) }
