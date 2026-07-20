//go:build windows
// +build windows

package svc

import "errors"

// privohubFilesystemQuota is unsupported on Windows; PrivoHub only ships the
// linux container. The caller falls back to the default tree-size quota.
func privohubFilesystemQuota(path string) (total, used, remaining uint64, err error) {
	return 0, 0, 0, errors.New("privohub filesystem quota not supported on windows")
}
