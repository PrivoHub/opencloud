//go:build !windows
// +build !windows

package svc

import "syscall"

// privohubFilesystemQuota returns df-style total/used/remaining bytes for the
// filesystem backing path. On PrivoHub the OpenCloud data dir and Immich uploads
// share a single ZFS dataset whose refquota is the plan size, so statfs reports
// the plan size as total and the combined Photos + Drive consumption as used.
func privohubFilesystemQuota(path string) (total, used, remaining uint64, err error) {
	var stat syscall.Statfs_t
	if err = syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0, err
	}
	bsize := uint64(stat.Bsize) //nolint:unconvert // Bsize is int64 on some platforms
	total = stat.Blocks * bsize
	// Bavail is the space available to unprivileged users; it is what the
	// refquota hard cap leaves free. Use it for both remaining and used so the
	// numbers stay consistent with what writes will actually be allowed.
	remaining = uint64(stat.Bavail) * bsize //nolint:unconvert // Bavail is int64 on freebsd
	used = total - remaining
	return total, used, remaining, nil
}
