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
