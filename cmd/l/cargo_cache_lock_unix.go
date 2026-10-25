//go:build darwin || linux

package main

import (
	"errors"
	"os"
	"syscall"

	"github.com/l-programming-language/l/lerrors"
)

func tryCargoCacheLock(path string) (func(), bool, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, false, lerrors.IO.Wrap(err, "open Rust cache publication lock")
	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == nil {
		return func() { _ = file.Close() }, true, nil //nolint:errcheck // Closing this unwritten lock descriptor releases ownership; the release callback has no error channel.
	}
	_ = file.Close() //nolint:errcheck // Preserve the flock error; this descriptor has no buffered output.
	if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
		return nil, false, nil
	}
	return nil, false, lerrors.IO.Wrap(err, "lock Rust build cache")
}
