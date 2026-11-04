package hardware_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

var devicesDB = testdb.Config{
	Prefix: "bulle_hardware_test",
	Migrate: func(db *gorm.DB) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return migrate.Run(context.Background(), migrate.Config{
			DB:     sqlDB,
			FS:     migrations.FS,
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			LockID: 7_300_313,
		})
	},
}

func openDevices(t *testing.T) *hardware.Postgres {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	db, err := testdb.Open(url, devicesDB)
	if err != nil {
		t.Fatal(err)
	}
	if err := testdb.Truncate(db, devicesDB); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{1, 2} {
		if err := db.Exec("INSERT INTO users (id, email) VALUES (?, ?)", id, fmt.Sprintf("u%d@facile.studio", id)).Error; err != nil {
			t.Fatal(err)
		}
	}
	return hardware.NewPostgres(db)
}

func TestPostgresKeepsDevicesToTheirOwner(t *testing.T) {
	devices := openDevices(t)
	ctx := context.Background()
	_, hash := hardware.MintToken()
	claimed, err := devices.Claim(ctx, hardware.Device{
		ID: hardware.NewID(), UserID: 1, Name: "Pastille #91C0", DeviceClass: "charm", HardwareID: "4F2A91C0",
	}, hash)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := devices.Owned(ctx, 2, claimed.ID); !errors.Is(err, hardware.ErrNotFound) {
		t.Fatalf("another user read the device: %v", err)
	}
	if err := devices.Revoke(ctx, 2, claimed.ID); !errors.Is(err, hardware.ErrNotFound) {
		t.Fatalf("another user revoked the device: %v", err)
	}
	if found, err := devices.ByTokenHash(ctx, hash); err != nil || found.ID != claimed.ID || found.Capabilities != "[]" {
		t.Fatalf("ByTokenHash = %+v, %v", found, err)
	}
	if err := devices.Revoke(ctx, 1, claimed.ID); err != nil {
		t.Fatalf("the owner could not revoke: %v", err)
	}
	if _, err := devices.ByTokenHash(ctx, hash); !errors.Is(err, hardware.ErrNotFound) {
		t.Fatalf("a revoked token still resolves: %v", err)
	}
}

func TestPostgresClaimingAgainReplacesTheToken(t *testing.T) {
	devices := openDevices(t)
	ctx := context.Background()
	_, oldHash := hardware.MintToken()
	_, newHash := hardware.MintToken()
	device := hardware.Device{ID: hardware.NewID(), UserID: 1, Name: "Desk", DeviceClass: "charm", HardwareID: "4F2A91C0"}
	first, err := devices.Claim(ctx, device, oldHash)
	if err != nil {
		t.Fatal(err)
	}
	device.ID = hardware.NewID()
	second, err := devices.Claim(ctx, device, newHash)
	if err != nil || second.ID != first.ID {
		t.Fatalf("re-claim = %+v, %v; want the same id %s", second, err, first.ID)
	}
	if _, err := devices.ByTokenHash(ctx, oldHash); !errors.Is(err, hardware.ErrNotFound) {
		t.Fatalf("the replaced token still resolves: %v", err)
	}
}
