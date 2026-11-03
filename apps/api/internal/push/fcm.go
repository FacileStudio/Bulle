package push

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

	// ChannelID is the Android notification channel the app creates at start.
	// The server and apps/client/src/lib/notifications.ts must agree on it.
	ChannelID = "bulle"

	fcmTTL = "86400s"
)

// ErrGone means the device no longer exists and should be forgotten.
var ErrGone = errors.New("push device is gone")

// FCM sends native Android notifications through Firebase Cloud Messaging.
type FCM struct {
	endpoint string
	client   *http.Client
}

// DecodeCredentials accepts a service account as raw JSON or base64 JSON.
func DecodeCredentials(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "{") {
		return []byte(value), nil
	}
	return base64.StdEncoding.DecodeString(value)
}

// NewFCM builds a sender from a Firebase service account key.
func NewFCM(ctx context.Context, credentials []byte) (*FCM, error) {
	creds, err := google.CredentialsFromJSON(ctx, credentials, fcmScope)
	if err != nil {
		return nil, fmt.Errorf("read the FCM service account: %w", err)
	}
	if creds.ProjectID == "" {
		return nil, errors.New("the FCM service account has no project_id")
	}
	return &FCM{
		endpoint: "https://fcm.googleapis.com/v1/projects/" + creds.ProjectID + "/messages:send",
		client:   &http.Client{Timeout: sendTimeout, Transport: &oauth2.Transport{Source: creds.TokenSource}},
	}, nil
}

type fcmMessage struct {
	Message fcmBody `json:"message"`
}

type fcmBody struct {
	Token        string            `json:"token"`
	Notification fcmNotification   `json:"notification"`
	Data         map[string]string `json:"data"`
	Android      fcmAndroid        `json:"android"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type fcmAndroid struct {
	Priority     string              `json:"priority"`
	TTL          string              `json:"ttl"`
	CollapseKey  string              `json:"collapse_key,omitempty"`
	Notification fcmAndroidNotifBody `json:"notification"`
}

type fcmAndroidNotifBody struct {
	ChannelID string `json:"channel_id"`
	Tag       string `json:"tag,omitempty"`
}

func fcmPayload(token string, n Notification) fcmMessage {
	return fcmMessage{Message: fcmBody{
		Token:        token,
		Notification: fcmNotification{Title: n.Title, Body: n.Body},
		Data:         map[string]string{"path": n.Path},
		Android: fcmAndroid{
			Priority:     "HIGH",
			TTL:          fcmTTL,
			CollapseKey:  n.Path,
			Notification: fcmAndroidNotifBody{ChannelID: ChannelID, Tag: n.Path},
		},
	}}
}

func (f *FCM) send(ctx context.Context, token string, n Notification) error {
	body, err := json.Marshal(fcmPayload(token, n))
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	reply, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusNotFound || bytes.Contains(reply, []byte("UNREGISTERED")) || badToken(resp.StatusCode, reply) {
		return ErrGone
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("fcm answered %d: %s", resp.StatusCode, bytes.TrimSpace(reply))
	}
	return nil
}

// badToken spots FCM's 400 for a malformed registration token. A 400 about the
// payload must not forget the device, so only the token field counts.
func badToken(status int, reply []byte) bool {
	return status == http.StatusBadRequest && bytes.Contains(reply, []byte("INVALID_ARGUMENT")) &&
		bytes.Contains(reply, []byte("registration token"))
}
