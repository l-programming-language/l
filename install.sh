#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$repo_dir"

if ! command -v go >/dev/null 2>&1; then
  printf 'Go is required. Install Go and ensure it is on PATH, then rerun this script.\n' >&2
  exit 1
fi

printf 'Downloading and verifying Go dependencies...\n'
go mod download
go mod verify

mkdir -p bin
go build -mod=readonly -o bin/l ./cmd/l

printf 'Built %s/bin/l with embedded language resources.\nThe compiler binary can be moved independently.\nRun "%s/bin/l" --help from any directory.\n' "$repo_dir" "$repo_dir"
