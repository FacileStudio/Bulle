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

// SearchHit is an artifact a search found. Match is title when the title
// holds every word, else content; Snippet is the body around the first word
// found, or the body's start for a title match.
type SearchHit struct {
	Artifact
	Match   string `json:"match"`
	Snippet string `json:"snippet"`
}

// SearchResponse is one page of a search, title matches first. NextCursor,
// when set, fetches the next page with the same q. Available is false, with
// no hits, when the user has no Mycelium connection.
type SearchResponse struct {
	Artifacts  []SearchHit `json:"artifacts"`
	NextCursor string      `json:"next_cursor,omitempty"`
	Available  bool        `json:"available"`
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
