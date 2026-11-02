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
	raw, err := c.gw.Get(ctx, userID, AppCourrier, fmt.Sprintf(courrierEmailPath, id, msg.ReplyTo), nil)
	if err != nil {
		return 0, sendRequest{}, err
	}
	var original courrierOriginal
	if err := decodeInto(raw, &original); err != nil {
		return 0, sendRequest{}, fmt.Errorf("courrier: reading email %d: %w", msg.ReplyTo, err)
	}
	return id, threaded(body, original), nil
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
