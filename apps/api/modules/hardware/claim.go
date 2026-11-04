package hardware

import (
	"regexp"
	"strings"
	"unicode/utf8"

	internal "github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const maxNameLength = 64

var hardwareIDPattern = regexp.MustCompile(`^[A-Z0-9:_-]{4,64}$`)

var deviceClasses = map[string]bool{"charm": true, "field": true, "totem": true}

func newDevice(userID int64, body ClaimRequest) (internal.Device, error) {
	hardwareID := strings.ToUpper(strings.TrimSpace(body.HardwareID))
	if !hardwareIDPattern.MatchString(hardwareID) {
		return internal.Device{}, troncerrors.Invalid("hardware_id must be 4 to 64 letters, digits, ':', '_' or '-'")
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
		name = "Pastille #" + hardwareID[len(hardwareID)-4:]
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
