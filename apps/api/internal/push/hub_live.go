package push

// Subscribe streams the user's live events to an open app.
func (h *Hub) Subscribe(userID int64) (<-chan Event, func()) {
	return h.cfg.Live.Subscribe(userID)
}

// Publish sends one live event to the user's open apps. A nil Hub drops it.
func (h *Hub) Publish(userID int64, eventType string, data any) {
	if h == nil {
		return
	}
	h.cfg.Live.Publish(userID, eventType, data)
}

// CloseStreams ends every open live stream and refuses new ones, for a
// server that is shutting down. A nil Hub has none.
func (h *Hub) CloseStreams() {
	if h != nil {
		h.cfg.Live.Close()
	}
}
