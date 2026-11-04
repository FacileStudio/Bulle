package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func serveMailboxes(rw http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case courrierAccountsPath:
		_, _ = fmt.Fprint(rw, `{"accounts":[{"id":7,"email":"me@work","is_default":true},{"id":8,"email":"me@ik.me"}]}`)
	case fmt.Sprintf(courrierSearchPath, 7):
		_, _ = fmt.Fprint(rw, `{"emails":[{"id":1,"subject":"old","date":"2026-10-01T09:00:00+02:00"}]}`)
	case fmt.Sprintf(courrierSearchPath, 8):
		_, _ = fmt.Fprint(rw, `{"emails":[{"id":42,"subject":"Visio de point","date":"2026-11-02T10:27:00+01:00"}]}`)
	case fmt.Sprintf(courrierEmailPath, 8, 42):
		_, _ = fmt.Fprint(rw, `{"message_id":"<m42@epitech>","subject":"Visio de point","from_address":"lisa@epitech.eu"}`)
	default:
		rw.WriteHeader(http.StatusNotFound)
	}
}

func TestSearchWithoutAccountCoversEveryMailbox(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.mailboxes = true
	raw, err := NewCourrier(NewGateway(w.tokens(), w.cfg, nil)).Search(context.Background(), 1, 0, "Nay", 10)
	if err != nil {
		t.Fatal(err)
	}
	var got mergedMail
	if err := json.Unmarshal(raw, &got); err != nil || len(got.Emails) != 2 {
		t.Fatalf("want both mailboxes merged, got %s, %v", raw, err)
	}
	first := got.Emails[0]
	if first["subject"] != "Visio de point" || first["account_id"] != float64(8) || first["account_email"] != "me@ik.me" {
		t.Fatalf("newest first, tagged with its mailbox: %v", first)
	}
}

func TestReplyFindsTheMailboxThatHoldsTheOriginal(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.mailboxes = true
	account, body, err := NewCourrier(NewGateway(w.tokens(), w.cfg, nil)).compose(context.Background(), 1, 7, CourrierMessage{ReplyTo: 42, Body: "relance"})
	if err != nil {
		t.Fatal(err)
	}
	if account != 8 || body.InReplyTo != "<m42@epitech>" || body.To[0] != "lisa@epitech.eu" || body.Subject != "Re: Visio de point" {
		t.Fatalf("reply must leave from mailbox 8 in the thread, got account %d, %+v", account, body)
	}
}
