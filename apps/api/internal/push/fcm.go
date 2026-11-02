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

const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

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
	Notification map[string]string `json:"notification"`
	Data         map[string]string `json:"data"`
	Android      map[string]string `json:"android"`
}

func (f *FCM) send(ctx context.Context, token string, n Notification) error {
	body, err := json.Marshal(fcmMessage{Message: fcmBody{
		Token:        token,
		Notification: map[string]string{"title": n.Title, "body": n.Body},
		Data:         map[string]string{"path": n.Path},
		Android:      map[string]string{"priority": "HIGH"},
	}})
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
	if resp.StatusCode == http.StatusNotFound || bytes.Contains(reply, []byte("UNREGISTERED")) {
		return ErrGone
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("fcm answered %d", resp.StatusCode)
	}
	return nil
}
