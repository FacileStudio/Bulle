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

func courrierTools(set *toolSet, c Courrier) {
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
}
