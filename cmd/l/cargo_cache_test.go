//go:build darwin || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/cargocache"
)

func TestCargoCacheAdmitsAvailableSlotAndCancelsWhenFull(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "Cargo.toml"), []byte("native input identity"), 0600); err != nil {
		t.Fatal(err)
	}
	var targets []string
	var releases []func()
	defer func() {
		for _, release := range releases {
			release()
		}
	}()
	for index := 0; index < 6; index++ {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		target, release, err := cargoBuildCache(ctx, source)
		cancel()
		if err != nil {
			t.Fatalf("free cache slot %d was not admitted: %v", index, err)
		}
		for _, occupied := range targets {
			if target == occupied {
				release()
				t.Fatal("two owners admitted to one cache slot")
			}
		}
		targets, releases = append(targets, target), append(releases, release)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	if _, release, err := cargoBuildCache(ctx, source); !errors.Is(err, context.DeadlineExceeded) {
		if release != nil {
			release()
		}
		t.Fatalf("full cache ignored bounded admission: %v", err)
	}
}

func TestCargoCachePublishesCurrentSourceUnderContention(t *testing.T) {
	const count = 8 // More owners than slots forces reuse as well as admission.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	directory := t.TempDir()
	var ready sync.WaitGroup
	ready.Add(count)
	errors := make(chan error, count)
	start := make(chan struct{})
	for index := 0; index < count; index++ {
		go func(index int) {
			ready.Done()
			<-start
			// Identical package and target names deliberately contend for the
			// same Cargo artifact paths. Their native results must stay distinct.
			manifest := []byte("[package]\nname = \"cache_witness\"\nversion = \"0.0.0\"\nedition = \"2024\"\n")
			if index%2 == 1 {
				manifest = append(manifest, []byte("publish = false\n")...)
			}
			// Keep this a same-shard contention witness after bounded sharding:
			// use the complete single-member workspace and native comments to
			// find distinct sources in the same bucket.
			probe := filepath.Join(directory, fmt.Sprintf("identity-%d", index))
			if err := os.MkdirAll(filepath.Join(probe, "orchestrator", "src"), 0700); err != nil {
				errors <- err
				return
			}
			var bundle *sourceBundle
			for nonce := 0; nonce < 256; nonce++ {
				source := []byte(fmt.Sprintf("fn main() { println!(\"source-%d\"); }\n// shard witness %d\n", index, nonce))
				candidate, err := planSourceBundle([][]byte{manifest, source}, []string{"Cargo.toml", "src/main.rs"})
				if err != nil {
					errors <- err
					return
				}
				lock, err := cargoWorkspaceLock([]*cargoPackage{candidate.packagePlan})
				if err != nil {
					errors <- err
					return
				}
				for name, data := range map[string][]byte{
					"Cargo.toml":               []byte("[workspace]\nresolver = \"3\"\nmembers = [\n \"orchestrator\",\n]\n"),
					"Cargo.lock":               lock,
					"orchestrator/Cargo.toml":  manifest,
					"orchestrator/src/main.rs": source,
				} {
					if err := os.WriteFile(filepath.Join(probe, name), data, 0600); err != nil {
						errors <- err
						return
					}
				}
				shard, err := cargocache.Shard(probe, "")
				if err != nil {
					errors <- err
					return
				}
				if shard == 0 {
					bundle = candidate
					break
				}
			}
			if bundle == nil {
				errors <- fmt.Errorf("could not construct same-shard native contention input")
				return
			}
			build := filepath.Join(directory, fmt.Sprintf("build-%d", index))
			output := filepath.Join(directory, fmt.Sprintf("app-%d", index))
			if err := compileCargoWorkspace(ctx, nil, bundle, nil, nil, build, nil, output); err != nil {
				errors <- err
				return
			}
			if shard, err := cargocache.Shard(build, filepath.Join(build, ".l-artifacts")); err != nil || shard != 0 {
				errors <- fmt.Errorf("actual workspace changed same-shard contention input: shard=%d: %v", shard, err)
				return
			}
			bytes, err := exec.CommandContext(ctx, output).CombinedOutput()
			if err != nil || string(bytes) != fmt.Sprintf("source-%d\n", index) {
				errors <- fmt.Errorf("artifact %d contains another build: %q: %w", index, bytes, err)
				return
			}
			errors <- nil
		}(index)
	}
	ready.Wait()
	close(start)
	for index := 0; index < count; index++ {
		if err := <-errors; err != nil {
			t.Error(err)
		}
	}
}

func TestCargoCacheLockWaitHonorsCancellation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "Cargo.toml"), []byte("native input identity"), 0600); err != nil {
		t.Fatal(err)
	}
	var locks []string
	for index := 0; index < cargoBuildSlots; index++ {
		target, release, err := cargoBuildCache(context.Background(), source)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(release)
		locks = append(locks, filepath.Join(filepath.Dir(target), "publication.lock"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	if _, release, err := cargoBuildCache(ctx, source); !errors.Is(err, context.DeadlineExceeded) {
		if release != nil {
			release()
		}
		t.Fatalf("contending cache lock ignored cancellation: %v", err)
	}
	for _, path := range locks {
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	}
}
