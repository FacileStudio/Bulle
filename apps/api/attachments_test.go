package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	attachmentsmodule "github.com/FacileStudio/Bulle/apps/api/modules/attachments"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/go-chi/chi/v5"
)

func TestAttachmentLoadsWithTheSessionCookie(t *testing.T) {
	manager, err := session.New(porte.Config{SuccessURL: "http://bulle.test/"}, session.Deps{
		Sessions: newMemoryStore(),
		Logger:   slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := manager.Issue(context.Background(), 7, "")
	if err != nil {
		t.Fatal(err)
	}
	st := store.NewMemory()
	created, _ := st.CreateAttachment(context.Background(), store.Attachment{ID: "img", UserID: 7, Name: "a.png", MediaType: "image/png", Size: 3, Data: []byte("png")})
	router := chi.NewRouter()
	attachmentsmodule.RegisterRoutes(router, st, middleware.RequireAuth(manager, fixedResolver{}))

	request := httptest.NewRequest(http.MethodGet, "/attachments/"+created.ID, nil)
	request.Header.Set("Cookie", porte.SessionCookieName+"="+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "png" || recorder.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("cookie GET = %d %q %q", recorder.Code, recorder.Body.String(), recorder.Header().Get("Content-Type"))
	}
}
