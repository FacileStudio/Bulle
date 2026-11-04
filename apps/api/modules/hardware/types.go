package hardware

import (
	"encoding/json"
	"time"
)

// ClaimRequest pairs a device. Name and DeviceClass are optional.
type ClaimRequest struct {
	HardwareID  string `json:"hardware_id"`
	Name        string `json:"name,omitempty"`
	DeviceClass string `json:"device_class,omitempty"`
}

// DeviceResponse is a paired device as the API shows it, never with its token.
type DeviceResponse struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	DeviceClass   string          `json:"device_class"`
	HardwareID    string          `json:"hardware_id"`
	Capabilities  json.RawMessage `json:"capabilities"`
	LastSeenAt    *time.Time      `json:"last_seen_at,omitempty"`
	LastTransport string          `json:"last_transport,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// ClaimResponse carries the device token, the only time the server returns it.
type ClaimResponse struct {
	Device       DeviceResponse `json:"device"`
	DeviceToken  string         `json:"device_token"`
	GatewayWSURL string         `json:"gateway_ws_url"`
}

// EmptyRequest documents a route that takes no body.
type EmptyRequest struct{}

// EmptyResponse documents a route that answers no body.
type EmptyResponse struct{}
