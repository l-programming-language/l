# Source-derived CLI help

Extracted mechanically from the actual `buildUsage` constant in [the command parser](../../cmd/l/command.go). This is **source-derived**, not runtime-captured.

```text
Usage: l build [--language LANGUAGE] [--assets ROOT] [--output-dir DIRECTORY] [main.l]

Build native macOS or Linux executables beside the source, or in --output-dir.
The default source is main.l. Options must precede the source filename.
Language resources are embedded in the executable. The default language is eng.
Use --assets to select a complete development resource bundle instead.
Cross-compilation and the legacy -os flag are not supported.
```
