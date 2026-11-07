package suite

import (
	"context"
	"fmt"
	"slices"
)

var courrierMarkActions = []string{"mark_read", "mark_unread", "star", "unstar"}

type courrierInboxInput struct {
	Account int64 `json:"account,omitempty" jsonschema:"description=Mail account id from courrier_list_accounts; omit to list every mailbox the user has"`
	Unread  bool  `json:"unread,omitempty" jsonschema:"description=Only unread conversations"`
	Limit   int   `json:"limit,omitempty" jsonschema:"description=Maximum results (default 10, max 50)"`
}

type courrierSearchInput struct {
	Query   string `json:"query" jsonschema:"required,description=Words to search for in subject, sender name, sender address and body; every word must match"`
	Account int64  `json:"account,omitempty" jsonschema:"description=Mail account id from courrier_list_accounts; omit to search every mailbox the user has"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=Maximum results (default 10, max 50)"`
}

type courrierAccountsInput struct{}

type courrierEmailInput struct {
	EmailID int64 `json:"email_id" jsonschema:"required,description=Email id as courrier_list_inbox or courrier_search answers it"`
	Account int64 `json:"account,omitempty" jsonschema:"description=The account_id the email was listed with; defaults to the user's default account"`
}

type courrierDraftRef struct {
	DraftID int64 `json:"draft_id" jsonschema:"required,description=Draft id as courrier_save_draft answers it"`
	Account int64 `json:"account,omitempty" jsonschema:"description=Mail account id the draft is in; defaults to the user's default account"`
}

type courrierSendInput struct {
	Account int64 `json:"account,omitempty" jsonschema:"description=Mail account to send from; when replying pass the account_id of the email you answer"`
	CourrierMessage
}

type courrierDraftInput struct {
	Account int64 `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	DraftID int64 `json:"draft_id,omitempty" jsonschema:"description=Draft id to overwrite; omit to create a new draft"`
	CourrierMessage
}

type courrierUpdateInput struct {
	Account  int64   `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	EmailIDs []int64 `json:"email_ids" jsonschema:"required,description=Email ids (max 200)"`
	Action   string  `json:"action" jsonschema:"required,enum=archive,enum=delete,description=delete moves to trash, or erases what is already in trash"`
}

type courrierMarkInput struct {
	Account  int64   `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	EmailIDs []int64 `json:"email_ids" jsonschema:"required,description=Email ids (max 200)"`
	Action   string  `json:"action" jsonschema:"required,enum=mark_read,enum=mark_unread,enum=star,enum=unstar"`
}

type courrierMoveInput struct {
	Account  int64   `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	EmailIDs []int64 `json:"email_ids" jsonschema:"required,description=Email ids (max 200)"`
	Folder   string  `json:"folder,omitempty" jsonschema:"enum=inbox,enum=archive,enum=trash,enum=junk,description=System folder to move to"`
	FolderID int64   `json:"folder_id,omitempty" jsonschema:"description=Custom folder id, instead of folder"`
}

func courrierTools(set *toolSet, c Courrier) {
	courrierReadTools(set, c)
	writeTool(set, "courrier_send", "Send an email from one of the user's Courrier mailboxes. To answer or follow up on an email, pass reply_to with that email's id and account with its account_id, so the message stays in the thread. Use this, not a draft, when the user asks to send, reply, follow up or relance: the user reviews and edits the email on an approval card before anything leaves.",
		func(ctx context.Context, in courrierSendInput) (string, error) {
			return rawResult(c.Send(ctx, set.userID, set.defaults.account(in.Account), in.CourrierMessage))
		})
	writeTool(set, "courrier_save_draft", "Save an email as a draft in Courrier without sending it, or overwrite the draft draft_id names. "+
		"Use it only when the user asks for a draft; to send or reply, use courrier_send.",
		func(ctx context.Context, in courrierDraftInput) (string, error) {
			return rawResult(c.SaveDraft(ctx, set.userID, set.defaults.account(in.Account), in.DraftID, in.CourrierMessage))
		})
	writeTool(set, "courrier_delete_draft", "Delete a Courrier draft that was never sent. draft_id comes from courrier_save_draft or the drafts folder. "+
		"It never touches a sent or received email; use courrier_update for those.",
		func(ctx context.Context, in courrierDraftRef) (string, error) {
			return done("deleted", c.DeleteDraft(ctx, set.userID, set.defaults.account(in.Account), in.DraftID))
		})
	courrierFlagTools(set, c)
}

func courrierFlagTools(set *toolSet, c Courrier) {
	selfTool(set, "courrier_mark", "Mark emails read or unread, or star or unstar them, in Courrier. Runs without an approval card. "+
		"Use it once the user has dealt with an email. action is one of mark_read, mark_unread, star or unstar.",
		func(ctx context.Context, in courrierMarkInput) (string, error) {
			if !slices.Contains(courrierMarkActions, in.Action) {
				return "", fmt.Errorf("courrier_mark: unknown action %q; valid actions are %v", in.Action, courrierMarkActions)
			}
			return done("done", c.Apply(ctx, set.userID, set.defaults.account(in.Account), in.EmailIDs, in.Action))
		})
	writeTool(set, "courrier_update", "Archive or delete emails in Courrier. delete moves them to trash, or erases what is already there. "+
		"To mark emails read or unread, or star them, use courrier_mark; to file them in a folder, use courrier_move.",
		func(ctx context.Context, in courrierUpdateInput) (string, error) {
			return done("done", c.Apply(ctx, set.userID, set.defaults.account(in.Account), in.EmailIDs, in.Action))
		})
	writeTool(set, "courrier_move", "Move emails to another Courrier folder: a system folder (inbox, archive, trash, junk) or a custom folder_id. "+
		"Pass exactly one of folder or folder_id.",
		func(ctx context.Context, in courrierMoveInput) (string, error) {
			return done("moved", c.Move(ctx, set.userID, set.defaults.account(in.Account), in.EmailIDs, CourrierFolder{Type: in.Folder, ID: in.FolderID}))
		})
}

func courrierReadTools(set *toolSet, c Courrier) {
	readTool(set, "courrier_list_inbox", "List the user's inbox conversations from Courrier, across every mailbox unless account is given. Each email carries account_id and account_email, and assistant_seen_at when an earlier run already showed it to you: the user has likely heard about that email, so lead with the others and bring it back only when it still needs the user's action.",
		func(ctx context.Context, in courrierInboxInput) (string, error) {
			return rawResult(set.markSeen(ctx)(c.Inbox(ctx, set.userID, set.defaults.account(in.Account), in.Unread, in.Limit)))
		})
	readTool(set, "courrier_search", "Search the user's emails in Courrier, across every mailbox unless account is given. Each email carries account_id and account_email. Search by a person's surname or company domain when the full name finds nothing.",
		func(ctx context.Context, in courrierSearchInput) (string, error) {
			return rawResult(set.markSeen(ctx)(c.Search(ctx, set.userID, set.defaults.account(in.Account), in.Query, in.Limit)))
		})
	readTool(set, "courrier_list_accounts", "List the user's own mailboxes in Courrier. These addresses belong to the user, never to a contact: when the user names one of them, they mean that mailbox.",
		func(ctx context.Context, _ courrierAccountsInput) (string, error) {
			return jsonResult(c.Accounts(ctx, set.userID))
		})
	readTool(set, "courrier_get_email", "Read one email from Courrier with its full text body and headers. "+
		"Use it when a listing's preview is not enough to answer. Pass the account_id the email was listed with as account.",
		func(ctx context.Context, in courrierEmailInput) (string, error) {
			return rawResult(c.Email(ctx, set.userID, set.defaults.account(in.Account), in.EmailID))
		})
}
