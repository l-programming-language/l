//go:build darwin || linux

package native

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Cargo locks its own products, but does not own our staged source inputs.
// Retain this inode and lock through staging, native execution and snapshots.
// Never unlink it: another process may already be waiting on the same inode.
func acquireBuildCache(ctx context.Context, target string) (func() error, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(target, 0700); err != nil {
		return nil, err
	}
	info, err := os.Lstat(target)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("Cargo target must be a real directory: %s", target)
	}
	// NONBLOCK prevents a preexisting nonregular path (for example a FIFO)
	// from hanging before the descriptor's regular-file check below.
	file, err := os.OpenFile(filepath.Join(target, ".l-parity-publication.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return nil, err
	}
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close() //nolint:errcheck // Preserve the unsafe-path or stat failure while releasing the unwritten descriptor.
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Cargo input lock must be a regular file")
	}
	for {
		if err := ctx.Err(); err != nil {
			_ = file.Close() //nolint:errcheck // Preserve cancellation while releasing the unwritten lock descriptor.
			return nil, err
		}
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			var once sync.Once
			var closeErr error
			return func() error { once.Do(func() { closeErr = file.Close() }); return closeErr }, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) && !errors.Is(err, syscall.EINTR) {
			_ = file.Close() //nolint:errcheck // Preserve the flock failure while releasing the unwritten descriptor.
			return nil, err
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			_ = file.Close() //nolint:errcheck // Preserve the deadline failure while releasing the unwritten descriptor.
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
