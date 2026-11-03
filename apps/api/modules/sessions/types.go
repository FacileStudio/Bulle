package sessions

import "time"

type SessionSummary struct {
	ID        string    `json:"id"`
	SpaceID   string    `json:"space_id"`
	Title     string    `json:"title"`
	Channel   string    `json:"channel,omitempty"`
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
	Answer   string           `json:"answer,omitempty"`
	Thinking string           `json:"thinking,omitempty"`
	Tools    []StoredToolCall `json:"tools,omitempty"`

	Attachments []AttachmentInfo `json:"attachments,omitempty"`
}

type AttachmentInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	Size      int64  `json:"size"`
}

type SessionDetailResponse struct {
	Session     SessionSummary  `json:"session"`
	Messages    []StoredMessage `json:"messages"`
	ActiveRunID string          `json:"active_run_id,omitempty"`
}

type CreateSessionRequest struct {
	Title   string `json:"title,omitempty"`
	SpaceID string `json:"space_id,omitempty"`
}

type CreateSessionResponse struct {
	ID      string `json:"id"`
	SpaceID string `json:"space_id"`
	Title   string `json:"title"`
}

type SendMessageRequest struct {
	Text        string   `json:"text"`
	Attachments []string `json:"attachments,omitempty"`
}

type SendMessageResponse struct {
	RunID string `json:"run_id"`
}

type DecideRequest struct {
	Approve bool `json:"approve"`
}

type EmptyResponse struct{}

type EmptyRequest struct{}
