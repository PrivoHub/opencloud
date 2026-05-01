package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencloud-eu/opencloud/services/idm/pkg/config"
)

func TestLoadAdminPasswordFile(t *testing.T) {
	t.Run("empty path is a no-op", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ServiceUserPasswords.OCAdmin = "from-yaml"
		if err := loadAdminPasswordFile(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ServiceUserPasswords.OCAdmin != "from-yaml" {
			t.Fatalf("OCAdmin should be untouched when path is empty, got %q", cfg.ServiceUserPasswords.OCAdmin)
		}
	})

	t.Run("missing file leaves OCAdmin untouched", func(t *testing.T) {
		dir := t.TempDir()
		cfg := &config.Config{}
		cfg.ServiceUserPasswords.OCAdmin = "from-yaml"
		cfg.ServiceUserPasswords.OCAdminPasswordFile = filepath.Join(dir, "does-not-exist")
		if err := loadAdminPasswordFile(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ServiceUserPasswords.OCAdmin != "from-yaml" {
			t.Fatalf("OCAdmin should be untouched when file is missing, got %q", cfg.ServiceUserPasswords.OCAdmin)
		}
	})

	t.Run("file value overrides yaml value", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "idm-admin")
		want := "deadbeefcafebabe1234567890abcdef"
		if err := os.WriteFile(path, []byte(want+"\n"), 0o600); err != nil {
			t.Fatalf("write file: %v", err)
		}
		cfg := &config.Config{}
		cfg.ServiceUserPasswords.OCAdmin = "from-yaml"
		cfg.ServiceUserPasswords.OCAdminPasswordFile = path
		if err := loadAdminPasswordFile(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ServiceUserPasswords.OCAdmin != want {
			t.Fatalf("OCAdmin should equal trimmed file contents, got %q want %q", cfg.ServiceUserPasswords.OCAdmin, want)
		}
	})

	t.Run("permission error is fatal", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("running as root, can't test permission error")
		}
		dir := t.TempDir()
		path := filepath.Join(dir, "idm-admin")
		if err := os.WriteFile(path, []byte("secret"), 0o000); err != nil {
			t.Fatalf("write file: %v", err)
		}
		cfg := &config.Config{}
		cfg.ServiceUserPasswords.OCAdminPasswordFile = path
		err := loadAdminPasswordFile(cfg)
		if err == nil {
			t.Fatal("expected error for unreadable file, got nil")
		}
		if !strings.Contains(err.Error(), "read admin password file") {
			t.Fatalf("error should mention the file, got: %v", err)
		}
	})
}
