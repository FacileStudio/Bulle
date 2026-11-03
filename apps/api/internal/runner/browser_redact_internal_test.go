package runner

import (
	"testing"

	"github.com/FacileStudio/Bulle/core/sandbox"
)

func TestRedactBodyHidesSecretFieldsAndKeepsTheRest(t *testing.T) {
	redact := NewRedactor("sk-live-123456789")
	cases := map[string]string{
		"email=a%40b.test&Password=hunter2&user%5Bpin%5D=1234&shipping=fast": "email=a%40b.test&Password=[redacted]&user%5Bpin%5D=[redacted]&shipping=fast",
		"cardNumber=4111&cc_cvv=123&otp=999&note=sk-live-123456789":          "cardNumber=[redacted]&cc_cvv=[redacted]&otp=[redacted]&note=[redacted]",
		`{"user":"ann","api_key":"abc","nested":{"accessToken":"t1"},"n":4}`: `{"user":"ann","api_key":"[redacted]","nested":{"accessToken":"[redacted]"},"n":4}`,
		`{"qty":2,"apiKey":42,"CVC":"12`:                                     `{"qty":2,"apiKey":"[redacted]","CVC":"[redacted]"`,
		`{"title":"pinned post","spinner":true}`:                             `{"title":"pinned post","spinner":true}`,
	}
	for body, want := range cases {
		if got := redactBody(body, redact); got != want {
			t.Errorf("redactBody(%q)\n got %q\nwant %q", body, got, want)
		}
	}
}

func TestHeldRecordRedactsTheBodyPreview(t *testing.T) {
	record := heldRecord(sandbox.HeldRequest{ID: "h1", Method: "POST", URL: "https://shop.test/login", BodyPreview: "user=ann&password=hunter2"}, nil)
	if record["body_preview"] != "user=ann&password=[redacted]" || record["url"] != "https://shop.test/login" {
		t.Fatalf("record = %v", record)
	}
}
