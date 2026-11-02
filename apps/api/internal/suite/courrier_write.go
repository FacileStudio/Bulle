package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

const (
	courrierEmailPath  = "/api/accounts/%d/mail/emails/%d"
	courrierSendPath   = "/api/accounts/%d/mail/send"
	courrierDraftsPath = "/api/accounts/%d/mail/drafts"
	courrierDraftPath  = "/api/accounts/%d/mail/drafts/%d"
	courrierBulkPath   = "/api/accounts/%d/mail/emails/bulk-action"
	courrierMovePath   = "/api/accounts/%d/mail/emails/move"
	maxCourrierIDs     = 200
)

var courrierEmailKeys = []string{
	"id", "folder_id", "message_id", "thread_id", "subject", "from_address", "from_name", "to_addresses",
	"cc_addresses", "date", "body_text", "is_read", "is_starred", "has_attachments", "attachments",
	"in_reply_to", "references", "category",
}

type CourrierMessage struct {
	To      []string `json:"to,omitempty" jsonschema:"description=Recipient addresses; defaults to the sender when replying"`
	Cc      []string `json:"cc,omitempty"`
	Bcc     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject,omitempty" jsonschema:"description=Subject; defaults to Re: <original subject> when replying"`
	Body    string   `json:"body" jsonschema:"required,description=Plain-text body"`
	ReplyTo int64    `json:"reply_to,omitempty" jsonschema:"description=Id of the email this answers, which threads the reply"`
}

type sendRequest struct {
	To         []string `json:"to"`
	Cc         []string `json:"cc,omitempty"`
	Bcc        []string `json:"bcc,omitempty"`
	Subject    string   `json:"subject"`
	Body       string   `json:"body"`
	InReplyTo  string   `json:"in_reply_to,omitempty"`
	References []string `json:"references,omitempty"`
}

type courrierOriginal struct {
	MessageID   string   `json:"message_id"`
	Subject     string   `json:"subject"`
	FromAddress string   `json:"from_address"`
	References  []string `json:"references"`
}

func (c Courrier) Email(ctx context.Context, userID, account, emailID int64) ([]byte, error) {
	id, err := c.account(ctx, userID, account)
	if err != nil {
		return nil, err
	}
	raw, err := c.gw.Get(ctx, userID, AppCourrier, fmt.Sprintf(courrierEmailPath, id, emailID), nil)
	if err != nil {
		return nil, err
	}
	return pickOne(raw, courrierEmailKeys...)
}

func (c Courrier) Send(ctx context.Context, userID, account int64, msg CourrierMessage) ([]byte, error) {
	id, body, err := c.compose(ctx, userID, account, msg)
	if err != nil {
		return nil, err
	}
	if len(body.To) == 0 {
		return nil, errors.New("an email needs at least one recipient")
	}
	return c.gw.Do(ctx, userID, Request{App: AppCourrier, Method: http.MethodPost, Path: fmt.Sprintf(courrierSendPath, id), Body: body})
}

func (c Courrier) SaveDraft(ctx context.Context, userID, account, draftID int64, msg CourrierMessage) ([]byte, error) {
	id, body, err := c.compose(ctx, userID, account, msg)
	if err != nil {
		return nil, err
	}
	if draftID > 0 {
		return c.gw.Do(ctx, userID, Request{App: AppCourrier, Method: http.MethodPut, Path: fmt.Sprintf(courrierDraftPath, id, draftID), Body: body})
	}
	return c.gw.Do(ctx, userID, Request{App: AppCourrier, Method: http.MethodPost, Path: fmt.Sprintf(courrierDraftsPath, id), Body: body})
}

func (c Courrier) DeleteDraft(ctx context.Context, userID, account, draftID int64) error {
	id, err := c.account(ctx, userID, account)
	if err != nil {
		return err
	}
	_, err = c.gw.Do(ctx, userID, Request{App: AppCourrier, Method: http.MethodDelete, Path: fmt.Sprintf(courrierDraftPath, id, draftID), Body: nil})
	return err
}

func (c Courrier) Apply(ctx context.Context, userID, account int64, emailIDs []int64, action string) error {
	id, err := c.batch(ctx, userID, account, emailIDs)
	if err != nil {
		return err
	}
	payload := map[string]any{"email_ids": emailIDs, "action": action}
	_, err = c.gw.Do(ctx, userID, Request{App: AppCourrier, Method: http.MethodPost, Path: fmt.Sprintf(courrierBulkPath, id), Body: payload})
	return err
}

type CourrierFolder struct {
	Type string
	ID   int64
}

func (c Courrier) Move(ctx context.Context, userID, account int64, emailIDs []int64, folder CourrierFolder) error {
	id, err := c.batch(ctx, userID, account, emailIDs)
	if err != nil {
		return err
	}
	payload := map[string]any{"email_ids": emailIDs}
	switch {
	case folder.ID > 0:
		payload["folder_id"] = folder.ID
	case folder.Type != "":
		payload["folder_type"] = folder.Type
	default:
		return errors.New("name the destination with folder or folder_id")
	}
	_, err = c.gw.Do(ctx, userID, Request{App: AppCourrier, Method: http.MethodPost, Path: fmt.Sprintf(courrierMovePath, id), Body: payload})
	return err
}

func (c Courrier) batch(ctx context.Context, userID, account int64, emailIDs []int64) (int64, error) {
	if len(emailIDs) == 0 || len(emailIDs) > maxCourrierIDs {
		return 0, fmt.Errorf("pass between 1 and %d email ids", maxCourrierIDs)
	}
	return c.account(ctx, userID, account)
}
