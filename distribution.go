// Package distribution contains the canonical resources shipped with L.
package distribution

import (
	"embed"
	"io/fs"
)

// Keep release assets explicit: developer settings, fixtures and unregistered
// build output must never become part of an installed compiler.
//
//go:embed lexicon/eng/grammar.json lexicon/eng/vocabulary.json
//go:embed builtins/eng/prelude.l builtins/eng/core.l
//go:embed runtime/eng/dictionary.l runtime/eng/foundations.l runtime/eng/errors.l
//go:embed runtime/eng/process.l runtime/eng/shutdown.l runtime/eng/app.l
//go:embed runtime/eng/start.l runtime/eng/supervision.l runtime/eng/capture.l
//go:embed runtime/generated/support.aml runtime/generated/support.json
//go:embed builtins/operations/rust/process-build.def builtins/operations/rust/process-run.def
//go:embed builtins/operations/rust/process-stop.def builtins/operations/rust/process-await.def
//go:embed builtins/operations/rust/process-finish-continuation.def
//go:embed dictionaries.json stdlib/eng/io/dictionary.l
//go:embed builtins/eng/storage/dictionary.l builtins/eng/storage/bindings.l builtins/eng/storage/values.l
//go:embed builtins/eng/storage/control.l builtins/eng/storage/display.l builtins/eng/storage/collections.l
//go:embed builtins/eng/storage/program.l builtins/eng/storage/traversal.l builtins/eng/storage/execution.l
//go:embed stdlib/eng/format/dictionary.l stdlib/eng/format/foundations.l
//go:embed builtins/segments/consonants/*.lu builtins/segments/vowels/*.lu
//go:embed builtins/templates/config/config.lu builtins/templates/install/cargo.lu
//go:embed builtins/templates/operate/start.lu builtins/templates/operate/stop.lu
//go:embed runtime/rust/main.rs runtime/rust/host.rs
//go:embed runtime/rust/process_unix.rs runtime/rust/process_other.rs
var resources embed.FS

// Files returns the immutable, embedded compiler distribution.
func Files() fs.FS { return resources }
