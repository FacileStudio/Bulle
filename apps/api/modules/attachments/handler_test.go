package attachments_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/modules/attachments"
	"github.com/go-chi/chi/v5"
)

func router(st store.Attachments) chi.Router {
	asUser := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var id int64 = 1
			if r.Header.Get("X-User") == "2" {
				id = 2
			}
			next.ServeHTTP(w, r.WithContext(authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: id})))
		})
	}
	r := chi.NewRouter()
	attachments.RegisterRoutes(r, st, asUser)
	return r
}

func uploadFile(t *testing.T, r http.Handler, name, browserType string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+name+`"`)
	if browserType != "" {
		header.Set("Content-Type", browserType)
	}
	part, err := form.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = form.Close()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/attachments", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestUploadDetectsTypeFromBytesAndName(t *testing.T) {
	r := router(store.NewMemory())
	cases := []struct {
		name, browser, want string
		data                string
	}{
		{"shot.png", "image/png", "image/png", "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"},
		{"doc.pdf", "", "application/pdf", "%PDF-1.7\n"},
		{"index.ts", "video/mp2t", "text/plain", "export const a: number = 1\n"},
		{"Dockerfile", "", "text/plain", "FROM alpine\n"},
		{"README.md", "application/octet-stream", "text/markdown", "# hi\n"},
	}
	for _, c := range cases {
		rec := uploadFile(t, r, c.name, c.browser, []byte(c.data))
		var got attachments.AttachmentResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		if rec.Code != http.StatusCreated || got.MediaType != c.want || got.Name != c.name || got.Size != int64(len(c.data)) || got.ID == "" {
			t.Errorf("%s: %d %+v", c.name, rec.Code, got)
		}
	}
}

func TestUploadRefusesSVGAndBinary(t *testing.T) {
	r := router(store.NewMemory())
	for name, data := range map[string]string{"logo.svg": "<svg xmlns='http://www.w3.org/2000/svg'/>", "a.bin": "\x00\x01\x02\xff"} {
		if rec := uploadFile(t, r, name, "", []byte(data)); rec.Code != http.StatusUnsupportedMediaType {
			t.Errorf("%s: %d", name, rec.Code)
		}
	}
}

func TestUploadSizeLimit(t *testing.T) {
	r := router(store.NewMemory())
	big := bytes.Repeat([]byte("a"), attachments.MaxSize+1)
	if rec := uploadFile(t, r, "big.txt", "", big); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("over the limit: %d %s", rec.Code, rec.Body.String())
	}
	if rec := uploadFile(t, r, "fits.txt", "", big[:attachments.MaxSize]); rec.Code != http.StatusCreated {
		t.Fatalf("at the limit: %d %s", rec.Code, rec.Body.String())
	}
}

func TestDownloadIsOwnerOnly(t *testing.T) {
	r := router(store.NewMemory())
	rec := uploadFile(t, r, "page.html", "text/html", []byte("<script>alert(1)</script>"))
	var created attachments.AttachmentResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	get := func(user string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/attachments/"+created.ID, nil)
		req.Header.Set("X-User", user)
		out := httptest.NewRecorder()
		r.ServeHTTP(out, req)
		return out
	}
	own := get("1")
	if own.Code != http.StatusOK || own.Body.String() != "<script>alert(1)</script>" {
		t.Fatalf("owner: %d %q", own.Code, own.Body.String())
	}
	if ct := own.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") || own.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("html must be served as plain text, got %q", ct)
	}
	if other := get("2"); other.Code != http.StatusNotFound {
		t.Fatalf("stranger: %d", other.Code)
	}
}
