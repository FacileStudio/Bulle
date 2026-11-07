package hardware

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}

var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}

var notFound = documentation.Error{Status: 404, Code: "not_found", Description: "The device does not exist or belongs to another account."}

var Documentation = documentation.Module{
	Name:        "hardware",
	Description: "BOHS-01 devices such as Pastille: pairing, revocation and the bohs.v1 gateway socket. Every route acts on the caller's devices only.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/hardware/devices",
			Summary:      "List paired devices",
			Description:  "Oldest first. Device tokens are never returned here.",
			Auth:         sessionAuth,
			ResponseBody: []DeviceResponse{},
			Errors:       []documentation.Error{unauthenticated},
		},
		{
			Method:       "POST",
			Path:         "/hardware/devices/claim",
			Summary:      "Pair a device to the caller",
			Description:  "Mints a bohs_tok_ device token and returns it once; the server keeps only its SHA-256. hardware_id is the 8 lowercase hex characters the device reports. Claiming a hardware_id again replaces its token and closes its live sockets. Two accounts can each pair the same hardware_id, since it is public in the device's advertising; the older pairing stops working once the device holds the newer token. An account pairs at most 16 devices, and claims are rate limited per user. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  ClaimRequest{},
			ResponseBody: ClaimResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body, hardware_id, name or device_class."},
				unauthenticated, csrf,
				{Status: 412, Code: "failed_precondition", Description: "The account already pairs 16 devices."},
				{Status: 429, Code: "rate_limited", Description: "Too many claims; retry after the Retry-After delay."},
			},
		},
		{
			Method:       "POST",
			Path:         "/hardware/devices/claim-code",
			Summary:      "Pair a remote device by its code",
			Description:  "For a device with no BLE link to the app, such as a Raspberry Pi in another city. The device shows the user_code it got from POST /hardware/pair; entering it pairs the device to the caller and queues its token for the device's next poll. Dashes, spaces and case are ignored. name overrides the name the device suggested. A code works once. Rate limited per user. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  ClaimCodeRequest{},
			ResponseBody: DeviceResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or name."},
				unauthenticated, csrf,
				{Status: 404, Code: "not_found", Description: "No device is waiting with this code, or it expired."},
				{Status: 412, Code: "failed_precondition", Description: "The account already pairs 16 devices."},
				{Status: 429, Code: "rate_limited", Description: "Too many claims; retry after the Retry-After delay."},
			},
		},
		{
			Method:       "POST",
			Path:         "/hardware/pair",
			Summary:      "Ask for a pairing code",
			Description:  "Called by the device itself, unauthenticated. Answers a user_code (XXXX-XXXX, valid 10 minutes) for the device to show, and a secret device_code it polls with every interval seconds. Rate limited per address.",
			Auth:         "none",
			RequestBody:  PairStartRequest{},
			ResponseBody: PairStartResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body, hardware_id, name or device_class."},
				{Status: 429, Code: "rate_limited", Description: "Too many pairing requests from this address."},
				{Status: 503, Code: "unavailable", Description: "Too many pairings in progress on the server."},
			},
		},
		{
			Method:       "POST",
			Path:         "/hardware/pair/poll",
			Summary:      "Wait for a pairing code to be entered",
			Description:  "Called by the device with its device_code. Answers status pending until a user enters the code, then claimed once with the device token and gateway URL, which the device keeps.",
			Auth:         "device_code in the body",
			RequestBody:  PairPollRequest{},
			ResponseBody: PairPollResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "The pairing expired, was already collected, or never existed."},
				{Status: 429, Code: "rate_limited", Description: "Polling too fast from this address."},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/hardware/devices/{id}",
			Summary:      "Revoke a device",
			Description:  "Deletes the device and closes its live gateway sockets with close code 4401. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  EmptyRequest{},
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound},
		},
		{
			Method:       "GET",
			Path:         "/hardware/ws",
			Summary:      "Open a bohs.v1 gateway socket",
			Description:  "WebSocket with subprotocol bohs.v1, one binary frame of at most 1024 bytes per message. A device on Wi-Fi authenticates with Authorization: Bearer bohs_tok_...; the signed-in app relaying over BLE or USB passes ?device_id=hw_... for a device it owns and ?transport=ble|usb (default ble). The server first sends a 0x02 tools/list request and gives the advertised tools to the owner's agent runs while the socket is open. It closes with 1009 on a larger frame, 1008 on a 0x03 provisioning frame, and 4401 when the device is revoked, its token replaced or the relaying session ends; it discards reserved opcodes. A device keeps at most two live sockets; a newer one replaces the oldest.",
			Auth:         "device bearer token, or session cookie or bearer token with device_id",
			ResponseBody: EmptyResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "The bohs.v1 subprotocol was not offered, or transport is not ble or usb."},
				{Status: 401, Code: "unauthenticated", Description: "No valid session or device token."},
				{Status: 403, Code: "forbidden", Description: "The relay request came from an origin that is not allowed."},
				notFound,
				{Status: 429, Code: "rate_limited", Description: "Too many socket upgrades from this address."},
			},
		},
	},
}
