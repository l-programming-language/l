//go:build !darwin && !linux

package native

import (
	"context"
	"fmt"
	"runtime"
)

func acquireBuildCache(context.Context, string) (func() error, error) {
	return nil, fmt.Errorf("native Cargo cache ownership requires macOS/Linux, got %s", runtime.GOOS)
}
