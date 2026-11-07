package computers

import (
	"context"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	internal "github.com/FacileStudio/Bulle/apps/api/internal/computers"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type Manager interface {
	Status(ctx context.Context, userID int64) (*internal.Computer, error)
	Launch(ctx context.Context, userID int64) (*internal.Computer, error)
	Stop(ctx context.Context, userID int64) (*internal.Computer, error)
	Snapshot(ctx context.Context, userID int64, tag string) (*internal.Computer, error)
	Rollback(ctx context.Context, userID int64, tag string) (*internal.Computer, error)
	SetNetwork(ctx context.Context, userID int64, network string) (*internal.Computer, error)
}

type Handler struct {
	manager Manager
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	h.reply(w, r, func(ctx context.Context, userID int64) (*internal.Computer, error) {
		return h.manager.Status(ctx, userID)
	})
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	h.reply(w, r, func(ctx context.Context, userID int64) (*internal.Computer, error) {
		return h.manager.Launch(ctx, userID)
	})
}

func (h *Handler) stop(w http.ResponseWriter, r *http.Request) {
	h.reply(w, r, func(ctx context.Context, userID int64) (*internal.Computer, error) {
		return h.manager.Stop(ctx, userID)
	})
}

func (h *Handler) snapshot(w http.ResponseWriter, r *http.Request) {
	h.tagged(w, r, h.manager.Snapshot)
}

func (h *Handler) rollback(w http.ResponseWriter, r *http.Request) {
	h.tagged(w, r, h.manager.Rollback)
}

func (h *Handler) network(w http.ResponseWriter, r *http.Request) {
	var body NetworkRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	h.reply(w, r, func(ctx context.Context, userID int64) (*internal.Computer, error) {
		return h.manager.SetNetwork(ctx, userID, body.Network)
	})
}

func (h *Handler) tagged(w http.ResponseWriter, r *http.Request, act func(context.Context, int64, string) (*internal.Computer, error)) {
	var body TagRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	h.reply(w, r, func(ctx context.Context, userID int64) (*internal.Computer, error) {
		return act(ctx, userID, body.Tag)
	})
}

func (h *Handler) reply(w http.ResponseWriter, r *http.Request, act func(context.Context, int64) (*internal.Computer, error)) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, errors.Unauthorized("not authenticated"))
		return
	}
	computer, err := act(r.Context(), identity.UserID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, toResponse(computer))
}

func toResponse(computer *internal.Computer) ComputerResponse {
	response := ComputerResponse{
		Status:   computer.Status,
		Error:    computer.Error,
		CPUs:     computer.CPUs,
		MemoryMB: computer.MemoryMB,
		DiskGB:   computer.DiskGB,
		Network:  computer.Network,
	}
	if computer.Status != internal.StatusAbsent && !computer.LastActiveAt.IsZero() {
		at := computer.LastActiveAt.UTC()
		response.LastActiveAt = &at
	}
	return response
}
