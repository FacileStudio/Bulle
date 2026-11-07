package memory

import "github.com/FacileStudio/Bulle/apps/api/internal/suite"

// ListResponse lists the memory pages of a space, profile.md first. Available
// is false, with an empty list, when the user has no Mycelium connection.
type ListResponse struct {
	Pages     []suite.MemoryPage `json:"pages"`
	Available bool               `json:"available"`
}

// PageResponse is one memory page with its markdown.
type PageResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteRequest replaces a memory page with new markdown.
type WriteRequest struct {
	Content string `json:"content"`
}
