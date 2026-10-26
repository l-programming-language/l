package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func stagedPublication(t *testing.T, parent, id string) (string, string) {
	t.Helper()
	stage, err := os.MkdirTemp(parent, "stage-")
	if err != nil {
		t.Fatal(err)
	}
	generation := filepath.Join(stage, "generation")
	if err := os.Mkdir(generation, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(generation, "app-000"), []byte(id), 0700); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(stage, "entry")
	if err := os.WriteFile(entry, []byte(id), 0700); err != nil {
		t.Fatal(err)
	}
	return generation, entry
}

func TestPublishGenerationPreservesPreviousOutput(t *testing.T) {
	for _, failure := range []int{1, 2} {
		t.Run([]string{"generation rename", "executable replacement"}[failure-1], func(t *testing.T) {
			dir := t.TempDir()
			output := filepath.Join(dir, "main")
			companion := output + ".apps"
			oldGeneration, oldEntry := stagedPublication(t, dir, "old")
			if err := publishGeneration(output, companion, "main", "old", oldGeneration, oldEntry, os.Rename); err != nil {
				t.Fatal(err)
			}
			generation, entry := stagedPublication(t, dir, "new")
			calls := 0
			sentinel := errors.New("injected rename failure")
			rename := func(from, to string) error {
				calls++
				if calls == failure {
					return sentinel
				}
				return os.Rename(from, to)
			}
			if err := publishGeneration(output, companion, "main", "new", generation, entry, rename); !errors.Is(err, sentinel) || !errors.Is(err, lerrors.IO) {
				t.Fatalf("error=%v", err)
			}
			if got, err := os.ReadFile(output); err != nil || string(got) != "old" {
				t.Fatalf("previous executable changed: %q %v", got, err)
			}
			if got, err := os.ReadFile(filepath.Join(companion, "generations", "old", "app-000")); err != nil || string(got) != "old" {
				t.Fatalf("previous App changed: %q %v", got, err)
			}
			if failure == 2 {
				if got, err := os.ReadFile(filepath.Join(companion, "generations", "new", "app-000")); err != nil || string(got) != "new" {
					t.Fatalf("completed unpublished generation not retained: %q %v", got, err)
				}
			}
		})
	}
}

func TestPublishConcurrentGenerationsRemainMatched(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "main")
	companion := output + ".apps"
	generation, entry := stagedPublication(t, dir, "old")
	if err := publishGeneration(output, companion, "main", "old", generation, entry, os.Rename); err != nil {
		t.Fatal(err)
	}
	type publication struct{ id, generation, entry string }
	var work []publication
	for _, id := range []string{"first", "second"} {
		generation, entry := stagedPublication(t, dir, id)
		work = append(work, publication{id, generation, entry})
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var group sync.WaitGroup
	for _, item := range work {
		item := item
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			results <- publishGeneration(output, companion, "main", item.id, item.generation, item.entry, os.Rename)
		}()
	}
	close(start)
	group.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	active, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	app, err := os.ReadFile(filepath.Join(companion, "generations", string(active), "app-000"))
	if err != nil || string(app) != string(active) {
		t.Fatalf("mixed publication: main=%q app=%q error=%v", active, app, err)
	}
	for _, id := range []string{"old", "first", "second"} {
		if _, err := os.Stat(filepath.Join(companion, "generations", id, "app-000")); err != nil {
			t.Fatalf("generation %s removed: %v", id, err)
		}
	}
}

func TestOutputRejectsUnrelatedDestinations(t *testing.T) {
	for _, kind := range []string{"file", "directory", "companion", "symlink", "generation symlink"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			output := filepath.Join(dir, "main")
			companion := output + ".apps"
			switch kind {
			case "file":
				if err := os.WriteFile(output, []byte("user data"), 0600); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(output, 0700); err != nil {
					t.Fatal(err)
				}
			case "companion":
				if err := os.Mkdir(companion, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Symlink(t.TempDir(), companion); err != nil {
					t.Fatal(err)
				}
			case "generation symlink":
				if err := os.Mkdir(companion, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(companion, ".l-owner"), []byte(companionOwner+"main\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(t.TempDir(), filepath.Join(companion, "generations")); err != nil {
					t.Fatal(err)
				}
			}
			if err := validateOutput(output, companion, "main"); err == nil {
				t.Fatal("unrelated destination accepted")
			}
			if kind == "file" {
				got, callErr := os.ReadFile(output)
				if callErr != nil {
					t.Fatal(callErr)
				}
				if !strings.EqualFold(string(got), "user data") {
					t.Fatal("existing file changed")
				}
			}
		})
	}
}

func TestInitialPublicationFailureLeavesNoIncompleteCompanion(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "main")
	generation, entry := stagedPublication(t, dir, "first")
	failure := errors.New("initial directory publication failed")
	err := publishGeneration(output, output+".apps", "main", "first", generation, entry, func(from, to string) error {
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("error=%v", err)
	}
	if _, err := os.Lstat(output + ".apps"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed initial publication left companion: %v", err)
	}
}
