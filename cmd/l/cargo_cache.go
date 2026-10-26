package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/l-programming-language/l/internal/cargocache"
	"github.com/l-programming-language/l/internal/timing"
	"github.com/l-programming-language/l/lerrors"
)

// Six bounded CLI slots can admit two native builds from each of three
// concurrent compiler clients. Test-helper sharding remains independent.
const cargoBuildSlots = 6

// Cargo owns dependency fingerprints. The coordinator only chooses a private,
// versioned cache and holds its publication lock until executables are copied
// into the current build's private staging directory.
func cargoBuildCache(ctx context.Context, source string) (target string, release func(), resultErr error) {
	identify := timing.Start("cache.identify", source)
	defer func() { identify(resultErr) }()
	shard, err := cargocache.Shard(source, filepath.Join(source, ".l-artifacts"))
	if err != nil {
		return "", nil, lerrors.IO.Wrap(err, "identify Rust build inputs")
	}
	identify(nil)
	wait := timing.Start("cache.wait", source)
	defer func() { wait(resultErr) }()
	base, err := os.UserCacheDir()
	if err != nil {
		return "", nil, lerrors.IO.Wrap(err, "locate Rust build cache")
	}
	root := filepath.Join(base, "l", "rust", nativeToolchainVersion, runtime.GOOS+"-"+runtime.GOARCH)
	for {
		if err := ctx.Err(); err != nil {
			return "", nil, lerrors.Tool.Wrap(err, "waiting for Rust build cache")
		}
		// Prefer the same input's usual cache for incremental reuse, but never
		// wait behind that slot while another bounded cache is available.
		for offset := 0; offset < cargoBuildSlots; offset++ {
			directory := filepath.Join(root, fmt.Sprintf("shard-%d", (shard+offset)%cargoBuildSlots))
			if err := os.MkdirAll(directory, 0700); err != nil {
				return "", nil, lerrors.IO.Wrap(err, "create Rust build cache")
			}
			info, err := os.Lstat(directory)
			if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return "", nil, lerrors.IO.Errorf("Rust build cache must be a directory: %s: %v", directory, err)
			}
			unlock, acquired, err := tryCargoCacheLock(filepath.Join(directory, "publication.lock"))
			if err != nil {
				return "", nil, err
			}
			if acquired {
				wait(nil)
				if timing.Enabled() {
					timing.Record("cache.slot", source, map[string]any{
						"preferred_slot": shard, "selected_slot": (shard + offset) % cargoBuildSlots,
						"target": filepath.Join(directory, "target"),
					})
				}
				held := timing.Start("cache.hold", directory)
				return filepath.Join(directory, "target"), func() { unlock(); held(nil) }, nil
			}
		}
		select {
		case <-ctx.Done():
			return "", nil, lerrors.Tool.Wrap(ctx.Err(), "waiting for Rust build cache")
		case <-time.After(25 * time.Millisecond):
		}
	}
}
