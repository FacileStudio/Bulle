package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	hardwaremodule "github.com/FacileStudio/Bulle/apps/api/modules/hardware"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/go-chi/chi/v5"
)

type stubDevices struct{}

func (stubDevices) Claim(_ context.Context, d hardware.Device, _ string) (hardware.Device, error) {
	return d, nil
}

func (stubDevices) Devices(context.Context, int64) ([]hardware.Device, error) { return nil, nil }

func (stubDevices) Revoke(context.Context, int64, string) error { return nil }

func (stubDevices) Owned(context.Context, int64, string) (hardware.Device, error) {
	return hardware.Device{}, hardware.ErrNotFound
}

func (stubDevices) ByTokenHash(context.Context, string) (hardware.Device, error) {
	return hardware.Device{}, hardware.ErrNotFound
}

func (stubDevices) Touch(context.Context, string, string, time.Time) error { return nil }

func hardwareRouter(t *testing.T) (chi.Router, string) {
	t.Helper()
	manager, err := session.New(porte.Config{SuccessURL: "http://bulle.test/"}, session.Deps{
		Sessions: newMemoryStore(),
		Logger:   slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	token, _, err := manager.Issue(context.Background(), 7, "")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	router := chi.NewRouter()
	hardwaremodule.RegisterRoutes(router, hardwaremodule.Config{
		Devices: stubDevices{}, Sockets: stubDevices{}, Registry: hardware.NewRegistry(),
	}, middleware.RequireAuth(manager, fixedResolver{}))
	return router, token
}

func TestHardwareMutationsNeedTheCSRFHeaderWithACookie(t *testing.T) {
	router, token := hardwareRouter(t)
	call := func(method, path string, csrf bool) int {
		request := httptest.NewRequest(method, path, strings.NewReader(`{"hardware_id":"4f2a91c0"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Cookie", porte.SessionCookieName+"="+token)
		if csrf {
			request.Header.Set(porte.CSRFHeaderName, "1")
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder.Code
	}
	want := map[bool]map[string]int{
		false: {http.MethodPost: http.StatusForbidden, http.MethodDelete: http.StatusForbidden},
		true:  {http.MethodPost: http.StatusCreated, http.MethodDelete: http.StatusNoContent},
	}
	paths := map[string]string{http.MethodPost: "/hardware/devices/claim", http.MethodDelete: "/hardware/devices/hw_x"}
	for csrf, codes := range want {
		for method, code := range codes {
			if got := call(method, paths[method], csrf); got != code {
				t.Errorf("%s %s with CSRF=%v = %d, want %d", method, paths[method], csrf, got, code)
			}
		}
	}
}
