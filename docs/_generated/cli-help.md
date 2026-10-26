# Source-derived CLI help

Extracted mechanically from the actual `buildUsage` constant in [the command parser](../../cmd/l/command.go). This is **source-derived**, not runtime-captured.

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
-os/--os is unsupported; target OS selection is planned after v0.0.1.
```
