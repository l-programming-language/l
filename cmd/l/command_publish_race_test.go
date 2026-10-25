package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestPublishConcurrentFirstGenerationsRemainMatched(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "main")
	companion := output + ".apps"
	type publication struct{ id, generation, entry string }
	var work []publication
	for _, id := range []string{"first", "second"} {
		generation, entry := stagedPublication(t, directory, id)
		work = append(work, publication{id, generation, entry})
	}
	start := make(chan struct{})
	results := make(chan error, len(work))
	var publishers sync.WaitGroup
	for _, item := range work {
		item := item
		publishers.Add(1)
		go func() {
			defer publishers.Done()
			<-start
			results <- publishGeneration(output, companion, "main", item.id, item.generation, item.entry, os.Rename)
		}()
	}
	close(start)
	publishers.Wait()
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
	if string(active) != "first" && string(active) != "second" {
		t.Fatalf("unknown active generation %q", active)
	}
	for _, item := range work {
		app, err := os.ReadFile(filepath.Join(companion, "generations", item.id, "app-000"))
		if err != nil || string(app) != item.id {
			t.Fatalf("generation %s does not match its executable: %q, %v", item.id, app, err)
		}
	}
}
