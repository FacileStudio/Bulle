package sessions

import "time"

type SessionSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StoredToolCall struct {
	CallID        string `json:"call_id"`
	Name          string `json:"name"`
	Args          any    `json:"args,omitempty"`
	Status        string `json:"status"`
	OutputPreview string `json:"output_preview,omitempty"`
	DurationMs    *int64 `json:"duration_ms,omitempty"`
}

type StoredMessage struct {
	Seq      int              `json:"seq"`
	Role     string           `json:"role"`
	Content  string           `json:"content"`
	Thinking string           `json:"thinking,omitempty"`
	Tools    []StoredToolCall `json:"tools,omitempty"`
}

type SessionDetailResponse struct {
	Session     SessionSummary  `json:"session"`
	Messages    []StoredMessage `json:"messages"`
	ActiveRunID string          `json:"active_run_id,omitempty"`
}

type CreateSessionRequest struct {
	Title string `json:"title,omitempty"`
}

type CreateSessionResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type SendMessageRequest struct {
	Text string `json:"text"`
}

type SendMessageResponse struct {
	RunID string `json:"run_id"`
}

type DecideRequest struct {
	Approve bool `json:"approve"`
}

type EmptyResponse struct{}

type EmptyRequest struct{}
