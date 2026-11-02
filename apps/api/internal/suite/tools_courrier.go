package suite

import "context"

type courrierInboxInput struct {
	Account int64 `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	Unread  bool  `json:"unread,omitempty" jsonschema:"description=Only unread conversations"`
	Limit   int   `json:"limit,omitempty" jsonschema:"description=Maximum results (default 10, max 50)"`
}

type courrierSearchInput struct {
	Query   string `json:"query" jsonschema:"required,description=Text to search for"`
	Account int64  `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=Maximum results (default 10, max 50)"`
}

type courrierAccountsInput struct{}

type courrierEmailInput struct {
	ID      int64 `json:"id" jsonschema:"required,description=Email id"`
	Account int64 `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
}

type courrierSendInput struct {
	Account int64 `json:"account,omitempty" jsonschema:"description=Mail account to send from; defaults to the user's default account"`
	CourrierMessage
}

type courrierDraftInput struct {
	Account int64 `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	ID      int64 `json:"id,omitempty" jsonschema:"description=Draft id to overwrite; omit to create a new draft"`
	CourrierMessage
}

type courrierUpdateInput struct {
	Account int64   `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	IDs     []int64 `json:"ids" jsonschema:"required,description=Email ids (max 200)"`
	Action  string  `json:"action" jsonschema:"required,enum=mark_read,enum=mark_unread,enum=star,enum=unstar,enum=archive,enum=delete,description=delete moves to trash, or erases what is already in trash"`
}

type courrierMoveInput struct {
	Account  int64   `json:"account,omitempty" jsonschema:"description=Mail account id; defaults to the user's default account"`
	IDs      []int64 `json:"ids" jsonschema:"required,description=Email ids (max 200)"`
	Folder   string  `json:"folder,omitempty" jsonschema:"enum=inbox,enum=archive,enum=trash,enum=junk,description=System folder to move to"`
	FolderID int64   `json:"folder_id,omitempty" jsonschema:"description=Custom folder id, instead of folder"`
}

func courrierTools(set *toolSet, c Courrier) {
	courrierReadTools(set, c)
	writeTool(set, "courrier_send", "Send an email from the user's Courrier account. Pass reply_to to answer an email in its thread.",
		func(ctx context.Context, in courrierSendInput) (string, error) {
			return rawResult(c.Send(ctx, set.userID, in.Account, in.CourrierMessage))
		})
	writeTool(set, "courrier_save_draft", "Save an email as a draft in Courrier without sending it, or overwrite an existing draft.",
		func(ctx context.Context, in courrierDraftInput) (string, error) {
			return rawResult(c.SaveDraft(ctx, set.userID, in.Account, in.ID, in.CourrierMessage))
		})
	writeTool(set, "courrier_delete_draft", "Delete a draft in Courrier.",
		func(ctx context.Context, in courrierEmailInput) (string, error) {
			return done("deleted", c.DeleteDraft(ctx, set.userID, in.Account, in.ID))
		})
	writeTool(set, "courrier_update", "Mark emails read or unread, star or unstar, archive or delete them in Courrier.",
		func(ctx context.Context, in courrierUpdateInput) (string, error) {
			return done("done", c.Apply(ctx, set.userID, in.Account, in.IDs, in.Action))
		})
	writeTool(set, "courrier_move", "Move emails to another Courrier folder.",
		func(ctx context.Context, in courrierMoveInput) (string, error) {
			return done("moved", c.Move(ctx, set.userID, in.Account, in.IDs, CourrierFolder{Type: in.Folder, ID: in.FolderID}))
		})
}

func courrierReadTools(set *toolSet, c Courrier) {
	readTool(set, "courrier_inbox", "List the user's inbox conversations from Courrier.",
		func(ctx context.Context, in courrierInboxInput) (string, error) {
			return rawResult(c.Inbox(ctx, set.userID, in.Account, in.Unread, in.Limit))
		})
	readTool(set, "courrier_search", "Search the user's emails in Courrier.",
		func(ctx context.Context, in courrierSearchInput) (string, error) {
			return rawResult(c.Search(ctx, set.userID, in.Account, in.Query, in.Limit))
		})
	readTool(set, "courrier_accounts", "List the user's mail accounts in Courrier.",
		func(ctx context.Context, _ courrierAccountsInput) (string, error) {
			return jsonResult(c.Accounts(ctx, set.userID))
		})
	readTool(set, "courrier_read", "Read one email from Courrier with its full text body and headers.",
		func(ctx context.Context, in courrierEmailInput) (string, error) {
			return rawResult(c.Email(ctx, set.userID, in.Account, in.ID))
		})
}
