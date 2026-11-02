package push_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

var devicesDB = testdb.Config{
	Prefix: "bulle_push_test",
	Migrate: func(db *gorm.DB) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return migrate.Run(context.Background(), migrate.Config{
			DB:     sqlDB,
			FS:     migrations.FS,
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			LockID: 7_300_302,
		})
	},
}

func TestPGDevicesMovesASharedTokenToTheNewOwner(t *testing.T) {
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
	ctx := context.Background()
	devices := push.NewPGDevices(db)
	if err := devices.SaveDevice(ctx, push.Device{UserID: 1, Kind: push.KindFCM, Token: "phone"}); err != nil {
		t.Fatal(err)
	}
	if err := devices.SaveDevice(ctx, push.Device{UserID: 2, Kind: push.KindFCM, Token: "phone"}); err != nil {
		t.Fatal(err)
	}
	first, _ := devices.Devices(ctx, 1)
	second, _ := devices.Devices(ctx, 2)
	if len(first) != 0 || len(second) != 1 || second[0].Kind != push.KindFCM {
		t.Fatalf("user 1 %+v, user 2 %+v", first, second)
	}
	if err := devices.DeleteDevice(ctx, 2, "phone"); err != nil {
		t.Fatal(err)
	}
	if left, _ := devices.Devices(ctx, 2); len(left) != 0 {
		t.Fatalf("device survived its delete: %+v", left)
	}
}
