package attachments

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

const (
	MaxSize       = 20 << 20
	multipartRoom = 64 << 10
	maxNameRunes  = 200
)

var errTooLarge = troncerrors.TooLarge("the file is larger than 20 MB")

type Handler struct {
	store store.Attachments
}

func NewHandler(st store.Attachments) *Handler { return &Handler{store: st} }

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxSize+multipartRoom)
	name, data, err := readFile(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	mediaType, ok := Detect(name, data)
	if !ok {
		httpjson.WriteJSON(w, http.StatusUnsupportedMediaType, map[string]any{
			"error": map[string]string{"code": "unsupported_media_type", "message": "only images, PDFs and text files can be attached"},
		})
		return
	}
	created, err := h.store.CreateAttachment(r.Context(), store.Attachment{
		ID: store.NewID(), UserID: identity.UserID, Name: name, MediaType: mediaType, Size: int64(len(data)), Data: data,
	})
	if errors.Is(err, store.ErrAttachmentQuota) {
		httpjson.WriteError(w, troncerrors.TooLarge("your attachments already use 1 GB, the most Bulle keeps per account"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to store the file", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, AttachmentResponse{ID: created.ID, Name: created.Name, MediaType: created.MediaType, Size: created.Size})
}

func readFile(r *http.Request) (string, []byte, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return "", nil, troncerrors.Invalid("expected a multipart body with a file field")
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			return "", nil, partError(err)
		}
		if part.FormName() != "file" {
			_ = part.Close()
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, MaxSize+1))
		_ = part.Close()
		if err != nil {
			return "", nil, partError(err)
		}
		if len(data) > MaxSize {
			return "", nil, errTooLarge
		}
		return cleanName(part.FileName()), data, nil
	}
}

func partError(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return errTooLarge
	}
	if errors.Is(err, io.EOF) {
		return troncerrors.Invalid("missing file field")
	}
	return troncerrors.Invalid("malformed multipart body")
}

func cleanName(name string) string {
	name = strings.TrimSpace(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
	if name == "" || name == "." || name == "/" {
		return "file"
	}
	if runes := []rune(name); len(runes) > maxNameRunes {
		return string(runes[:maxNameRunes])
	}
	return name
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	found, err := h.store.GetAttachment(r.Context(), identity.UserID, chi.URLParam(r, "id"))
	if errors.Is(err, store.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.NotFound("attachment not found"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to load the file", err))
		return
	}
	header := w.Header()
	header.Set("Content-Type", served(found.MediaType))
	header.Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": found.Name}))
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	header.Set("Cache-Control", "private, max-age=31536000, immutable")
	_, _ = w.Write(found.Data)
}
