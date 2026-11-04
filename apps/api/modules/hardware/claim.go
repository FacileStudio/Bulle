package hardware

import (
	"regexp"
	"strings"
	"unicode/utf8"

	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const maxNameLength = 64

var hardwareIDPattern = regexp.MustCompile(`^[0-9a-f]{8}$`)

var deviceClasses = map[string]bool{"charm": true, "field": true, "totem": true}

func newDevice(userID int64, body ClaimRequest) (internal.Device, error) {
	hardwareID := body.HardwareID
	if !hardwareIDPattern.MatchString(hardwareID) {
		return internal.Device{}, troncerrors.Invalid("hardware_id must be 8 lowercase hex characters")
	}
	class := strings.TrimSpace(body.DeviceClass)
	if class == "" {
		class = "charm"
	}
	if !deviceClasses[class] {
		return internal.Device{}, troncerrors.Invalid("device_class must be charm, field or totem")
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "Pastille #" + strings.ToUpper(hardwareID[4:])
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return internal.Device{}, troncerrors.Invalid("name must be at most 64 characters")
	}
	return internal.Device{
		ID:          internal.NewID(),
		UserID:      userID,
		Name:        name,
		DeviceClass: class,
		HardwareID:  hardwareID,
	}, nil
}
