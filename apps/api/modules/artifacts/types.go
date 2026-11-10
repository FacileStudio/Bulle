package artifacts

import "time"

// Artifact is one Mycelium artifact the user can read. MyceliumSpace is the
// tree it lives in, "common" for an admin's personal common tree, and
// SpaceName that space's name, empty for the common tree. Machine is the
// machine or app that wrote it, such as bulle or ruche.
type Artifact struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Machine       string    `json:"machine"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
	MyceliumSpace string    `json:"mycelium_space"`
	SpaceName     string    `json:"space_name"`
	SessionID     string    `json:"session_id,omitempty"`
}

// ListResponse lists every artifact the user can read, newest first. Available is false, with an empty
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

// UpdateRequest is the body of PUT /artifacts/{id}: the whole new markdown
// and, when given, a new title.
type UpdateRequest struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content"`
}
