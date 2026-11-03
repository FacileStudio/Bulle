package chat

import (
	"testing"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func roomMessage(content *event.MessageEventContent) *event.Event {
	return &event.Event{
		Type: event.EventMessage, Sender: "@me:m.org", RoomID: "!room:m.org", ID: "$ask",
		Content: event.Content{Parsed: content},
	}
}

func TestMessageFromSeesMentionsAndThreads(t *testing.T) {
	const self = id.UserID("@bot:m.org")
	plain, _ := messageFrom(roomMessage(&event.MessageEventContent{MsgType: event.MsgText, Body: "lunch?"}), self)
	if plain.Mentioned || plain.Thread != "$ask" {
		t.Errorf("plain = %+v, want no mention and a thread rooted at the message", plain)
	}
	pill := &event.MessageEventContent{MsgType: event.MsgText, Body: "Bot: deploy", Mentions: &event.Mentions{UserIDs: []id.UserID{self}}}
	if msg, _ := messageFrom(roomMessage(pill), self); !msg.Mentioned {
		t.Error("m.mentions naming the bot did not count as a mention")
	}
	typed := &event.MessageEventContent{MsgType: event.MsgText, Body: "@bot:m.org deploy"}
	if msg, _ := messageFrom(roomMessage(typed), self); !msg.Mentioned {
		t.Error("the bot's user id in the body did not count as a mention")
	}
	inThread := &event.MessageEventContent{MsgType: event.MsgText, Body: "and staging?", RelatesTo: (&event.RelatesTo{}).SetThread("$root", "$prev")}
	if msg, _ := messageFrom(roomMessage(inThread), self); msg.Thread != "$root" || msg.Key() != "matrix:!room:m.org:$root" {
		t.Errorf("thread message = %+v, want thread $root", msg)
	}
}

func TestReplyGoesIntoTheThread(t *testing.T) {
	content := replyTo(Identity{Room: "!room:m.org", Sender: "@me:m.org", Event: "$ask", Thread: "$root"}, "done")
	if content.RelatesTo.GetThreadParent() != "$root" || content.RelatesTo.GetReplyTo() != "$ask" {
		t.Errorf("relates_to = %+v, want thread $root replying to $ask", content.RelatesTo)
	}
}
