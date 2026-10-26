//go:build !darwin && !linux

package main

import (
	"github.com/l-programming-language/l/lerrors"
)

func tryCargoCacheLock(string) (func(), bool, error) {
	return nil, false, lerrors.Unsupported.Errorf("native Rust compilation requires macOS or Linux")
}
