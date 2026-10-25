package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestBuildDirectoryCleanup(t *testing.T) {
	owned := filepath.Join(t.TempDir(), "owned-staging")
	buildCause := errors.New("build failed")
	primary := lerrors.Tool.Wrap(buildCause, "build failed")
	otherCause := errors.New("earlier cleanup failed")
	joined := errors.Join(primary, otherCause)
	cleanup := &os.PathError{Op: "remove", Path: owned, Err: os.ErrPermission}
	for _, tc := range []struct {
		name    string
		primary error
		cleanup error
	}{
		{"success", nil, nil},
		{"published success survives cleanup failure", nil, cleanup},
		{"original failure survives successful cleanup", primary, nil},
		{"original joined failure survives successful cleanup", joined, nil},
		{"build and cleanup failures survive", primary, cleanup},
		{"existing joined and cleanup failures survive", joined, cleanup},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			err := cleanupBuildDirectory(owned, tc.primary, func(directory string) error {
				calls++
				if directory != owned {
					t.Errorf("removed %q, want owned directory %q", directory, owned)
				}
				return tc.cleanup
			})
			if calls != 1 {
				t.Errorf("remove calls = %d, want 1", calls)
			}
			if tc.primary == nil || tc.cleanup == nil {
				if err != tc.primary {
					t.Errorf("cleanup changed original result identity: got %v, want %v", err, tc.primary)
				}
				return
			}
			for _, cause := range []error{tc.primary, buildCause, lerrors.Tool, cleanup, os.ErrPermission, lerrors.IO} {
				if !errors.Is(err, cause) {
					t.Errorf("cleanup result %v lost cause %v", err, cause)
				}
			}
			if tc.primary == joined && !errors.Is(err, otherCause) {
				t.Errorf("cleanup result %v lost the existing joined cause", err)
			}
			var pathError *os.PathError
			if !errors.As(err, &pathError) || pathError != cleanup {
				t.Errorf("cleanup result lost typed filesystem cause: got %#v, want %#v", pathError, cleanup)
			}
		})
	}
}

type cargoArtifactTestCloser struct {
	calls int
	err   error
}

func (c *cargoArtifactTestCloser) Close() error {
	c.calls++
	return c.err
}

func TestCargoArtifactClose(t *testing.T) {
	artifactCause := errors.New("artifact rejected")
	primary := lerrors.Tool.Wrap(artifactCause, "artifact rejected")
	otherCause := errors.New("copy failed")
	joined := errors.Join(primary, otherCause)
	closeFailure := &os.PathError{Op: "close", Path: filepath.Join(t.TempDir(), "artifact"), Err: os.ErrPermission}
	for _, tc := range []struct {
		name    string
		primary error
		close   error
	}{
		{"success", nil, nil},
		{"original failure survives successful close", primary, nil},
		{"original joined failure survives successful close", joined, nil},
		{"close failure rejects artifact", nil, closeFailure},
		{"artifact and close failures survive", primary, closeFailure},
		{"existing joined and close failures survive", joined, closeFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			closer := &cargoArtifactTestCloser{err: tc.close}
			err := closeCargoArtifact(closer, tc.primary)
			if closer.calls != 1 {
				t.Errorf("close calls = %d, want 1", closer.calls)
			}
			if tc.close == nil {
				if err != tc.primary {
					t.Errorf("close changed original result identity: got %v, want %v", err, tc.primary)
				}
				return
			}
			for _, cause := range []error{closeFailure, os.ErrPermission, lerrors.IO} {
				if !errors.Is(err, cause) {
					t.Errorf("close result %v lost cause %v", err, cause)
				}
			}
			if tc.primary != nil {
				for _, cause := range []error{tc.primary, artifactCause, lerrors.Tool} {
					if !errors.Is(err, cause) {
						t.Errorf("close result %v lost original cause %v", err, cause)
					}
				}
			}
			if tc.primary == joined && !errors.Is(err, otherCause) {
				t.Errorf("close result %v lost the existing joined cause", err)
			}
			var pathError *os.PathError
			if !errors.As(err, &pathError) || pathError != closeFailure {
				t.Errorf("close result lost typed filesystem cause: got %#v, want %#v", pathError, closeFailure)
			}
		})
	}
}
