package memory

import (
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Handler proxies a Bulle space's memory pages to its Mycelium memory space
// with the user's own delegated token, so the user sees, edits and deletes
// exactly what the agent's memory tools read and write.
type Handler struct {
	store   store.Store
	gateway *suite.Gateway
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, space, err := h.resolve(r)
	if err == nil {
		var pages []suite.MemoryPage
		pages, err = suite.NewMycelium(h.gateway).ListPages(r.Context(), userID, space)
		if err == nil {
			httpjson.WriteJSON(w, http.StatusOK, ListResponse{Pages: pages, Available: true})
			return
		}
	}
	if disconnected(err) {
		httpjson.WriteJSON(w, http.StatusOK, ListResponse{Pages: []suite.MemoryPage{}, Available: false})
		return
	}
	httpjson.WriteError(w, routeError(err))
}

func (h *Handler) read(w http.ResponseWriter, r *http.Request) {
	page, err := h.target(r)
	var content []byte
	if err == nil {
		content, err = suite.NewMycelium(h.gateway).Read(r.Context(), page.userID, page.space, page.escaped)
	}
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, PageResponse{Path: page.path, Content: string(content)})
}

func (h *Handler) write(w http.ResponseWriter, r *http.Request) {
	var body WriteRequest
	if err := httpjson.DecodeJSONLimit(w, r, &body, 2*suite.MaxMemoryPage); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if strings.TrimSpace(body.Content) == "" || len(body.Content) > suite.MaxMemoryPage {
		httpjson.WriteError(w, troncerrors.Invalid("content must be between 1 byte and 64 KB"))
		return
	}
	page, err := h.target(r)
	if err == nil {
		err = suite.NewMycelium(h.gateway).Write(r.Context(), page.userID, page.space, page.escaped, body.Content)
	}
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, PageResponse{Path: page.path, Content: body.Content})
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	page, err := h.target(r)
	if err == nil {
		err = suite.NewMycelium(h.gateway).Delete(r.Context(), page.userID, page.space, page.escaped)
	}
	if err != nil {
		httpjson.WriteError(w, routeError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
