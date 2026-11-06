package artifacts

import "time"

// Artifact is one Mycelium artifact of a Bulle space.
type Artifact struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Machine   string    `json:"machine"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	SessionID string    `json:"session_id,omitempty"`
}

// ListResponse lists a space's artifacts. Available is false, with an empty
// list, when the user has no Mycelium connection.
type ListResponse struct {
	Artifacts []Artifact `json:"artifacts"`
	Available bool       `json:"available"`
}

// DetailResponse is an artifact with its markdown, frontmatter removed.
type DetailResponse struct {
	Artifact
	Content string `json:"content"`
}
