package suite

import (
	"context"
	"fmt"
	"strings"
)

func (c Courrier) compose(ctx context.Context, userID, account int64, msg CourrierMessage) (int64, sendRequest, error) {
	id, err := c.account(ctx, userID, account)
	if err != nil {
		return 0, sendRequest{}, err
	}
	body := sendRequest{To: msg.To, Cc: msg.Cc, Bcc: msg.Bcc, Subject: msg.Subject, Body: msg.Body}
	if msg.ReplyTo <= 0 {
		return id, body, nil
	}
	id, raw, err := c.original(ctx, userID, id, msg.ReplyTo)
	if err != nil {
		return 0, sendRequest{}, err
	}
	var original courrierOriginal
	if err := decodeInto(raw, &original); err != nil {
		return 0, sendRequest{}, fmt.Errorf("courrier: reading email %d: %w", msg.ReplyTo, err)
	}
	return id, threaded(body, original), nil
}

// original reads the email being answered. When it is not in the given
// account, the other mailboxes are tried, and the reply leaves from the one
// that holds it, so it stays in the thread.
func (c Courrier) original(ctx context.Context, userID, account, emailID int64) (int64, []byte, error) {
	raw, first := c.gw.Get(ctx, userID, AppCourrier, fmt.Sprintf(courrierEmailPath, account, emailID), nil)
	if first == nil {
		return account, raw, nil
	}
	accounts, err := c.Accounts(ctx, userID)
	if err != nil {
		return 0, nil, first
	}
	for _, acc := range accounts {
		if acc.ID == account {
			continue
		}
		if raw, err := c.gw.Get(ctx, userID, AppCourrier, fmt.Sprintf(courrierEmailPath, acc.ID, emailID), nil); err == nil {
			return acc.ID, raw, nil
		}
	}
	return 0, nil, fmt.Errorf("courrier: email %d to reply to is in none of the user's mailboxes: %w", emailID, first)
}

func threaded(body sendRequest, original courrierOriginal) sendRequest {
	if len(body.To) == 0 && original.FromAddress != "" {
		body.To = []string{original.FromAddress}
	}
	if body.Subject == "" {
		body.Subject = original.Subject
		if !strings.HasPrefix(strings.ToLower(body.Subject), "re:") {
			body.Subject = "Re: " + body.Subject
		}
	}
	if original.MessageID != "" {
		body.InReplyTo = original.MessageID
		body.References = append(append([]string{}, original.References...), original.MessageID)
	}
	return body
}
