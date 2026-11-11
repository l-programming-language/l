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
//go:embed builtins/eng/core.l
//go:embed builtins/operations/rust/process-build.def builtins/operations/rust/process-run.def
//go:embed builtins/operations/rust/process-stop.def builtins/operations/rust/process-await.def
//go:embed builtins/operations/rust/process-finish-continuation.def
//go:embed dictionaries.json
//go:embed builtins/segments/consonants/*.lu builtins/segments/vowels/*.lu
//go:embed builtins/templates/config/config.lu builtins/templates/install/cargo.lu
//go:embed builtins/templates/operate/start.lu builtins/templates/operate/stop.lu
var resources embed.FS

// Files returns the immutable, embedded compiler distribution.
func Files() fs.FS { return resources }
