package env

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// hardwareGatewayURL reads BULLE_HARDWARE_GATEWAY_URL, the socket URL handed
// to paired devices. Empty means wss:// on the request host; ws:// is for a
// local development setup only.
func hardwareGatewayURL() (string, error) {
	value := strings.TrimSpace(os.Getenv("BULLE_HARDWARE_GATEWAY_URL"))
	if err := validateGatewayURL(value); err != nil {
		return "", err
	}
	return value, nil
}

func validateGatewayURL(value string) error {
	if value == "" {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "wss" && parsed.Scheme != "ws") || parsed.Host == "" {
		return fmt.Errorf("BULLE_HARDWARE_GATEWAY_URL must be a wss:// URL (ws:// for local development only)")
	}
	return nil
}
