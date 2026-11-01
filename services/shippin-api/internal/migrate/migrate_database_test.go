package migrate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Uses a new temporary database, not the database holding the CI fixtures.
func TestCustomerAccessMigrationPreservesExistingData(t *testing.T) {
	url := os.Getenv("SHIPPIN_TEST_DISPOSABLE_DATABASE_URL")
	if url == "" {
		t.Skip("disposable database not configured")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	name := "shippin_migration_test_" + hex.EncodeToString(b)
	quoted := pgx.Identifier{name}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+quoted); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(ctx, "DROP DATABASE "+quoted+" WITH (FORCE)"); err != nil {
			t.Errorf("drop temporary database: %v", err)
		}
	})
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.Database = name
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate source tree")
	}
	fullDir := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../data/schema/sql"))
	oldDir := t.TempDir()
	migrations, err := ListDir(fullDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version >= "0009" {
			continue
		}
		contents, err := os.ReadFile(migration.Path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(oldDir, filepath.Base(migration.Path)), contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Up(ctx, db, oldDir); err != nil {
		t.Fatalf("prepare old schema: %v", err)
	}
	var seeded int
	if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM shippin_core.subjects)
		+ (SELECT count(*) FROM shippin_core.tenants) + (SELECT count(*) FROM shippin_infra.public_ips)
		+ (SELECT count(*) FROM shippin_infra.platforms)`).Scan(&seeded); err != nil {
		t.Fatal(err)
	}
	if seeded != 0 {
		t.Fatal("fresh migrations must not create installation-specific rows")
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO shippin_core.tenants(slug,name) VALUES ('migration-fixture','Existing');
		INSERT INTO shippin_core.subjects(kind,display_name,email,status,email_verified_at)
		  VALUES ('customer','Existing','migration-fixture@example.test','active',now());
		INSERT INTO shippin_infra.public_ips(address,role,status) VALUES ('192.0.2.41','ingress','active');
		INSERT INTO shippin_infra.platforms(stable_id,kind) VALUES ('platform.migration-fixture','docker');`); err != nil {
		t.Fatal(err)
	}
	ran, err := Up(ctx, db, fullDir)
	if err != nil {
		t.Fatalf("upgrade schema: %v", err)
	}
	if len(ran) != 1 || ran[0] != "0009_customer_access" {
		t.Fatalf("unexpected migrations: %v", ran)
	}
	var count int
	err = db.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM shippin_core.tenants WHERE slug='migration-fixture') +
		(SELECT count(*) FROM shippin_core.subjects WHERE email='migration-fixture@example.test') +
		(SELECT count(*) FROM shippin_infra.public_ips WHERE address='192.0.2.41'::inet) +
		(SELECT count(*) FROM shippin_infra.platforms WHERE stable_id='platform.migration-fixture')`).Scan(&count)
	if err != nil || count != 4 {
		t.Fatalf("existing records preserved: count=%d error=%v", count, err)
	}
	ran, err = Up(ctx, db, fullDir)
	if err != nil || len(ran) != 0 {
		t.Fatalf("rerun: applied=%v error=%v", ran, err)
	}
}
