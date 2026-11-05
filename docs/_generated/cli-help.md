# Source-derived CLI help

Extracted mechanically from the actual usage constants below. This is **source-derived**, not runtime-captured.

## Build and top-level help

From `buildUsage` in [cmd/l/command.go](../../cmd/l/command.go).

```text
Usage: l build [--output-dir DIRECTORY] [main.l]
       l (--help | -h | -help)
       l build (--help | -h | -help | --h)
       l (--version | -v)

Build for the native macOS/Linux host; defaults: main.l, output beside source.
Help prints usage; version prints the embedded major.minor.patch version and newline.
Top-level help/version flags stand alone. Build options/help precede the filename.
Use -- to end option parsing. English resources are embedded.
Developer overrides: --language LANGUAGE (default: eng); --assets ROOT (complete bundle).
Optional native provenance: --native-evidence-output PATH records tools, compilation, and artifacts.
-os/--os is unsupported; target OS selection is planned after v0.0.1.
```

## Contracts help

From `contractsUsage` in [cmd/l/contracts.go](../../cmd/l/contracts.go).

```text
Usage: l contracts inspect [--format text|json] [--output PATH] INPUT
       l contracts check --profile NAME [--profile NAME ...] [--report PATH] INPUT
       l contracts check --request-file PATH [--native-environment PATH] [--report PATH] INPUT
       l contracts evidence capture --output PATH [--target HOST_TRIPLE] INPUT
       l contracts explain --binding OCCURRENCE_OR_SOURCE:BYTE_OFFSET INPUT
       l contracts verify-dictionary [--report PATH] DICTIONARY
       l contracts verify-dictionary --assurance symbolic --adjective MEANING_OR_SOURCE:BYTE_OFFSET [--adjective ...] DICTIONARY
Options precede input. Common options: --assets ROOT, --language eng,
--manifest PATH (repeatable), --trust-policy PATH, --native-environment PATH,
--format text|json. Inspection and checking never execute external tools.
Explicit evidence capture queries installed tools; it never builds or runs Apps.
Track A preserves current resolution. Track B has no source-facing mode.
Assurance defaults to implementation. Symbolic assurance checks only selected
source adjective definitions; native/runtime assurance remains separate.
```
