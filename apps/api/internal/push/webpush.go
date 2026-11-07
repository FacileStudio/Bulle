package push

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// WebPush sends browser notifications signed with the server's VAPID keys.
type WebPush struct {
	publicKey  string
	privateKey string
	subject    string
}

// NewWebPush returns a sender, or nil when either VAPID key is missing.
func NewWebPush(publicKey, privateKey, subject string) *WebPush {
	if publicKey == "" || privateKey == "" {
		return nil
	}
	return &WebPush{publicKey: publicKey, privateKey: privateKey, subject: subject}
}

func (w *WebPush) send(ctx context.Context, d Device, n Notification) error {
	payload, err := json.Marshal(map[string]string{"title": n.Title, "body": n.Body, "path": n.Path})
	if err != nil {
		return err
	}
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: d.Token,
		Keys:     webpush.Keys{P256dh: d.P256dh, Auth: d.Auth},
	}, &webpush.Options{
		HTTPClient:      &http.Client{Timeout: sendTimeout},
		Subscriber:      w.subject,
		VAPIDPublicKey:  w.publicKey,
		VAPIDPrivateKey: w.privateKey,
		TTL:             3600,
		Urgency:         webpush.UrgencyHigh,
	})
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return ErrGone
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("web push answered %d", resp.StatusCode)
	}
	return nil
}
