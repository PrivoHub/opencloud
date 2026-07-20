//go:build !windows
// +build !windows

package svc

import (
	"os"
	"testing"
)

func TestPrivohubFilesystemQuota(t *testing.T) {
	total, used, remaining, err := privohubFilesystemQuota(os.TempDir())
	if err != nil {
		t.Fatalf("statfs on temp dir failed: %v", err)
	}
	if total == 0 {
		t.Fatal("expected non-zero total for a real filesystem")
	}
	if used+remaining != total {
		t.Fatalf("used (%d) + remaining (%d) should equal total (%d)", used, remaining, total)
	}
	if remaining > total {
		t.Fatalf("remaining (%d) must not exceed total (%d)", remaining, total)
	}
}

func TestPrivohubFilesystemQuotaMissingPath(t *testing.T) {
	if _, _, _, err := privohubFilesystemQuota("/privohub/does/not/exist"); err == nil {
		t.Fatal("expected error for a nonexistent path")
	}
}
