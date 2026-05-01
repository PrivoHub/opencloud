package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/opencloud-eu/opencloud/services/idm/pkg/config"
)

// loadAdminPasswordFile, when cfg.ServiceUserPasswords.OCAdminPasswordFile is set and the file
// exists, replaces cfg.ServiceUserPasswords.OCAdmin with the trimmed file contents. A missing
// file is not an error (vanilla / first-boot path); any other read error is fatal so a
// misconfigured deployment does not silently fall back to a stale password.
func loadAdminPasswordFile(cfg *config.Config) error {
	path := cfg.ServiceUserPasswords.OCAdminPasswordFile
	if path == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read admin password file %q: %w", path, err)
	}
	cfg.ServiceUserPasswords.OCAdmin = strings.TrimSpace(string(b))
	return nil
}
