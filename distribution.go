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
//go:embed stdlib/eng/runtime/dictionary.l stdlib/eng/runtime/foundations.l stdlib/eng/runtime/errors.l
//go:embed stdlib/eng/runtime/process.l stdlib/eng/runtime/shutdown.l stdlib/eng/runtime/app.l
//go:embed stdlib/eng/runtime/start.l stdlib/eng/runtime/supervision.l stdlib/eng/runtime/capture.l
//go:embed stdlib/runtime/generated/support.aml stdlib/runtime/generated/support.json
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
//go:embed stdlib/runtime/rust/main.rs stdlib/runtime/rust/host.rs
//go:embed stdlib/runtime/rust/process_unix.rs stdlib/runtime/rust/process_other.rs
var resources embed.FS

// Files returns the immutable, embedded compiler distribution.
func Files() fs.FS { return resources }
