package hardware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/FacileStudio/Bulle/apps/api/modules/hardware"
)

func decode[T any](t *testing.T, res *http.Response, status int) T {
	t.Helper()
	var out T
	if res.StatusCode != status {
		t.Fatalf("status = %d, want %d", res.StatusCode, status)
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestCodePairingHandsTheTokenToTheDeviceOnce(t *testing.T) {
	st := newFakeStore()
	srv := server(t, st)
	start := decode[hardware.PairStartResponse](t,
		send(t, http.MethodPost, srv.URL+"/hardware/pair", "", `{"hardware_id":"b8e1c0de","device_class":"field","name":"Living Room Light"}`),
		http.StatusCreated)
	if len(start.UserCode) != 9 || start.UserCode[4] != '-' || start.DeviceCode == "" {
		t.Fatalf("start = %+v", start)
	}
	poll := `{"device_code":"` + start.DeviceCode + `"}`
	if got := decode[hardware.PairPollResponse](t, send(t, http.MethodPost, srv.URL+"/hardware/pair/poll", "", poll), http.StatusOK); got.Status != "pending" || got.DeviceToken != "" {
		t.Fatalf("poll before the code = %+v", got)
	}
	typed := strings.ToLower(strings.ReplaceAll(start.UserCode, "-", " "))
	device := decode[hardware.DeviceResponse](t,
		send(t, http.MethodPost, srv.URL+"/hardware/devices/claim-code", "1", `{"code":"`+typed+`"}`),
		http.StatusCreated)
	if device.Name != "Living Room Light" || device.DeviceClass != "field" || device.HardwareID != "b8e1c0de" {
		t.Fatalf("claimed %+v", device)
	}
	if res := send(t, http.MethodPost, srv.URL+"/hardware/devices/claim-code", "2", `{"code":"`+start.UserCode+`"}`); res.StatusCode != http.StatusNotFound {
		t.Fatalf("a second user entering the same code got %d", res.StatusCode)
	}
	got := decode[hardware.PairPollResponse](t, send(t, http.MethodPost, srv.URL+"/hardware/pair/poll", "", poll), http.StatusOK)
	if got.Status != "claimed" || !internal.IsToken(got.DeviceToken) || got.Device == nil || got.Device.ID != device.ID {
		t.Fatalf("poll after the code = %+v", got)
	}
	owner, err := st.ByTokenHash(context.Background(), internal.HashToken(got.DeviceToken))
	if err != nil || owner.UserID != 1 {
		t.Fatalf("the token does not open the device of user 1: %+v, %v", owner, err)
	}
	if res := send(t, http.MethodPost, srv.URL+"/hardware/pair/poll", "", poll); res.StatusCode != http.StatusNotFound {
		t.Fatalf("a second poll got %d, the token must leave the server once", res.StatusCode)
	}
}

func TestCodePairingRefusesAnUnknownCode(t *testing.T) {
	srv := server(t, newFakeStore())
	if res := send(t, http.MethodPost, srv.URL+"/hardware/devices/claim-code", "1", `{"code":"AAAA-AAAA"}`); res.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown code got %d", res.StatusCode)
	}
	if res := send(t, http.MethodPost, srv.URL+"/hardware/pair", "", `{"hardware_id":"NOTHEX"}`); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad hardware_id got %d", res.StatusCode)
	}
}
