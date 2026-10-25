//go:build !darwin && !linux

package main

import "errors"

func lockDependencyCache(string) (func() error, error) {
	return nil, errors.New("Rust dependency cache preparation supports macOS and Linux")
}
