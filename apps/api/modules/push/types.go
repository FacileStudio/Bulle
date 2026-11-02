package push

// StatusResponse says which channels deliver notifications. Server, Topic and
// SubscribeURL describe the ntfy topic and are empty when ntfy is off.
type StatusResponse struct {
	Enabled      bool   `json:"enabled"`
	Server       string `json:"server"`
	Topic        string `json:"topic"`
	SubscribeURL string `json:"subscribe_url"`
	Native       bool   `json:"native"`
	WebPushKey   string `json:"web_push_key,omitempty"`
}

// DeviceRequest registers a device: an FCM token (kind "fcm"), or a browser
// subscription endpoint with its keys (kind "web").
type DeviceRequest struct {
	Kind   string `json:"kind"`
	Token  string `json:"token"`
	P256dh string `json:"p256dh,omitempty"`
	Auth   string `json:"auth,omitempty"`
}

// ForgetDeviceRequest names the device to stop notifying.
type ForgetDeviceRequest struct {
	Token string `json:"token"`
}

type EmptyRequest struct{}

type EmptyResponse struct{}
