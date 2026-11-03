package chat

import (
	"context"

	"maunium.net/go/mautrix/id"
)

// inline turns a message sent outside any thread in a one-to-one room into
// an inline one. A thread reply keeps its thread, so a person who opens one
// in a direct message is still answered there.
func (m *Matrix) inline(ctx context.Context, msg Message) Message {
	if msg.Thread != msg.Event || !m.direct(ctx, id.RoomID(msg.Room)) {
		return msg
	}
	msg.Thread, msg.Inline = "", true
	return msg
}

// direct reports whether exactly two people can read the room: the bot and
// one other. The state store answers from the sync; an empty cache asks the
// homeserver once.
func (m *Matrix) direct(ctx context.Context, roomID id.RoomID) bool {
	if m.client == nil {
		return false
	}
	if m.client.StateStore != nil {
		if members, err := m.client.StateStore.GetRoomJoinedOrInvitedMembers(ctx, roomID); err == nil && len(members) > 0 {
			return len(members) == 2
		}
	}
	joined, err := m.client.JoinedMembers(ctx, roomID)
	return err == nil && len(joined.Joined) == 2
}
