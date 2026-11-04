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
			Description:  "Mints a bohs_tok_ device token and returns it once; the server keeps only its SHA-256. Claiming a hardware_id again replaces its token and closes its live sockets. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  ClaimRequest{},
			ResponseBody: ClaimResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body, hardware_id, name or device_class."},
				unauthenticated, csrf,
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
			Description:  "WebSocket with subprotocol bohs.v1, one binary frame per message. A device on Wi-Fi authenticates with Authorization: Bearer bohs_tok_...; the signed-in app relaying over BLE passes ?device_id=hw_... for a device it owns. The server first sends a 0x02 tools/list request.",
			Auth:         "device bearer token, or session cookie or bearer token with device_id",
			ResponseBody: EmptyResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "The bohs.v1 subprotocol was not offered."},
				{Status: 401, Code: "unauthenticated", Description: "No valid session or device token."},
				{Status: 403, Code: "forbidden", Description: "The relay request came from an origin that is not allowed."},
				notFound,
			},
		},
	},
}
