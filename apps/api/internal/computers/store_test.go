package computers_test

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/computers"
	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
)

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason(""))
	}
	config := testdb.Config{
		Prefix: "bulle_test",
		Migrate: func(db *gorm.DB) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return migrate.Run(context.Background(), migrate.Config{DB: sqlDB, FS: migrations.FS})
		},
	}
	db, err := testdb.Open(url, config)
	if err != nil {
		t.Fatal(err)
	}
	if err := testdb.Truncate(db, config); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestGormStoreRoundTrip(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	if err := db.Exec(`INSERT INTO users (id, email) VALUES (1, 'a@b.c'), (2, 'd@e.f')`).Error; err != nil {
		t.Fatal(err)
	}
	store := computers.NewStore(db)

	if found, err := store.Find(ctx, 1); err != nil || found != nil {
		t.Fatalf("found = %v err = %v", found, err)
	}
	old := time.Now().Add(-time.Hour)
	if err := store.Save(ctx, &computers.Computer{UserID: 1, BoxID: "b1", Status: computers.StatusRunning, SSHPort: 2222, LastActiveAt: old}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ctx, &computers.Computer{UserID: 1, BoxID: "b1", Status: computers.StatusRunning, SSHPort: 2223, LastActiveAt: old}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := store.Save(ctx, &computers.Computer{UserID: 2, BoxID: "b2", Status: computers.StatusRunning, LastActiveAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	found, err := store.Find(ctx, 1)
	if err != nil || found == nil || found.SSHPort != 2223 {
		t.Fatalf("found = %+v err = %v", found, err)
	}
	idle, err := store.Idle(ctx, time.Now().Add(-time.Minute))
	if err != nil || len(idle) != 1 || idle[0].UserID != 1 {
		t.Fatalf("idle = %+v err = %v", idle, err)
	}
	if err := store.Touch(ctx, 1, time.Now()); err != nil {
		t.Fatal(err)
	}
	if idle, _ := store.Idle(ctx, time.Now().Add(-time.Minute)); len(idle) != 0 {
		t.Fatalf("touched row still idle: %+v", idle)
	}
	if err := store.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if found, _ := store.Find(ctx, 1); found != nil {
		t.Fatal("row survived delete")
	}
}
