//go:build darwin || linux

package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// Use the same flock inode as the CLI/helper. Never unlink the lock file:
// another process may already be waiting on that inode.
func lockDependencyCache(path string) (func() error, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		if err == nil {
			err = fmt.Errorf("dependency cache lock must be regular: %s", path)
		}
		return nil, errors.Join(err, file.Close())
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, errors.Join(fmt.Errorf("dependency cache is busy or cannot be locked; finish other builds before bootstrap: %s: %w", path, err), file.Close())
	}
	return file.Close, nil
}
