package push_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/push"
)

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestNotifyPostsToTheUserTopic(t *testing.T) {
	var path, title, click, auth, body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		path, body = r.URL.Path, string(raw)
		title, click, auth = r.Header.Get("Title"), r.Header.Get("Click"), r.Header.Get("Authorization")
	}))
	defer server.Close()

	sender := push.New(push.Config{URL: server.URL + "/", Token: "tk", TopicSecret: "s"}, quiet())
	sender.Notify(context.Background(), 7, push.Notification{Title: "Bulle", Body: "digest is ready", Path: "sessions/abc"})

	if path != "/"+sender.Topic(7) || title != "Bulle" || body != "digest is ready" {
		t.Fatalf("path=%q title=%q body=%q", path, title, body)
	}
	if click != "studio.facile.bulle://sessions/abc" || auth != "Bearer tk" {
		t.Fatalf("click=%q auth=%q", click, auth)
	}
}

func TestTopicIsStableAndPerUser(t *testing.T) {
	a := push.New(push.Config{URL: "http://x", TopicSecret: "s"}, quiet())
	b := push.New(push.Config{URL: "http://x", TopicSecret: "other"}, quiet())
	topic := a.Topic(1)
	if topic != a.Topic(1) || topic == a.Topic(2) || topic == b.Topic(1) {
		t.Fatal("topic must be stable, per user and per secret")
	}
	if !strings.HasPrefix(topic, "bulle-") || len(topic) != len("bulle-")+32 {
		t.Fatalf("unexpected topic %q", topic)
	}
}

func TestDisabledSenderIsANoOp(t *testing.T) {
	var none *push.Sender
	none.Notify(context.Background(), 1, push.Notification{})
	if none.Enabled() || push.New(push.Config{URL: "http://x"}, nil) != nil || push.New(push.Config{TopicSecret: "s"}, nil) != nil {
		t.Fatal("missing url or secret must disable push")
	}
}

func TestUpstreamFailureIsSwallowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	url := server.URL
	sender := push.New(push.Config{URL: url, TopicSecret: "s"}, quiet())
	sender.Notify(context.Background(), 1, push.Notification{Title: "Bulle", Body: "x", Path: "p"})
	server.Close()
	sender.Notify(context.Background(), 1, push.Notification{Title: "Bulle", Body: "x", Path: "p"})
}
