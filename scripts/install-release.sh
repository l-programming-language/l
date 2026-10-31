#!/usr/bin/env bash
# Install a published L archive. This script never builds the compiler from source.
set -euo pipefail
export LC_ALL=C

fail() { printf 'L installation failed: %s\n' "$*" >&2; exit 1; }
usage() {
  cat <<'USAGE'
Usage: bash install-release.sh [--version VERSION|latest] [--repository OWNER/REPO]
                              [--prefix DIRECTORY] [--with-toolchain|--binary-only]
                              [--from-dir DIRECTORY]

Install the matching Apple Silicon macOS/Linux binary into ~/.local/bin (or --prefix).
By default, prepare dependencies and ask before installing missing Rust tools.
--with-toolchain permits installing missing Rust tools without a prompt.
--binary-only installs just l; native builds still need the documented prerequisites.
No sudo is used, and shell profiles are never modified.
--from-dir reads local release assets with the same verification; --version is required.
USAGE
}

release_version=latest
release_repository=l-programming-language/l
release_prefix=${HOME:?HOME must identify your user directory}/.local/bin
release_setup_mode=interactive
release_from_dir=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --version|--repository|--prefix|--from-dir)
      [ "$#" -ge 2 ] && [ -n "$2" ] || fail "Missing value for $1"
      case "$1" in
        --version) release_version=${2#v} ;;
        --repository) release_repository=$2 ;;
        --prefix) release_prefix=$2 ;;
        --from-dir) release_from_dir=$2 ;;
      esac
      shift 2 ;;
    --with-toolchain)
      [ "$release_setup_mode" = interactive ] || fail 'Choose only one setup mode'
      release_setup_mode=with-toolchain; shift ;;
    --binary-only)
      [ "$release_setup_mode" = interactive ] || fail 'Choose only one setup mode'
      release_setup_mode=binary-only; shift ;;
    --help|-h) usage; exit 0 ;;
    *) fail "Unknown argument: $1" ;;
  esac
done
[[ "$release_repository" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || fail 'Invalid GitHub repository'
[[ "$release_version" = latest || "$release_version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[A-Za-z0-9]+([.-][A-Za-z0-9]+)*)?$ ]] || fail 'Invalid release version'
if [ -n "$release_from_dir" ]; then
  [ "$release_version" != latest ] || fail '--from-dir requires an explicit --version'
  [ -d "$release_from_dir" ] || fail 'Local release asset directory does not exist'
  release_from_dir=$(cd -- "$release_from_dir" && pwd)
fi

case "$(uname -s)" in Darwin) release_os=darwin ;; Linux) release_os=linux ;; *) fail 'Supported systems are macOS and Linux' ;; esac
case "$(uname -m)" in arm64|aarch64) release_arch=arm64 ;; x86_64|amd64) release_arch=amd64 ;; *) fail 'Supported CPUs are ARM64 and x86-64' ;; esac
[[ "$release_os-$release_arch" != darwin-amd64 ]] || fail 'macOS releases require a native Apple Silicon (arm64) shell; Intel Macs are not supported'
for release_command in curl tar sort awk cmp mktemp; do command -v "$release_command" >/dev/null 2>&1 || fail "Required command is missing: $release_command"; done
if command -v sha256sum >/dev/null 2>&1; then
  digest() { sha256sum "$1" | awk '{print $1}'; }
elif command -v shasum >/dev/null 2>&1; then
  digest() { shasum -a 256 "$1" | awk '{print $1}'; }
else
  fail 'SHA-256 verification requires sha256sum or shasum'
fi
download() { curl --proto '=https' --proto-redir '=https' --tlsv1.2 --fail --silent --show-error --location --retry 3 "$@"; }

release_work=$(mktemp -d "${TMPDIR:-/tmp}/l-install.XXXXXXXX")
release_replacement=
cleanup() {
  if [ -n "$release_replacement" ]; then rm -f -- "$release_replacement"; fi
  rm -rf -- "$release_work"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

if [ "$release_version" = latest ]; then
  release_latest=$(download --head --output /dev/null --write-out '%{url_effective}' "https://github.com/$release_repository/releases/latest")
  release_tag_prefix="https://github.com/$release_repository/releases/tag/v"
  [[ "$release_latest" = "$release_tag_prefix"* ]] || fail 'Latest release did not resolve to a versioned GitHub release'
  release_version=${release_latest#"$release_tag_prefix"}
  [[ "$release_version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[A-Za-z0-9]+([.-][A-Za-z0-9]+)*)?$ ]] || fail 'Latest release has an invalid version'
fi

release_asset="l_${release_version}_${release_os}_${release_arch}.tar.gz"
release_url="https://github.com/$release_repository/releases/download/v$release_version"
printf 'Downloading L %s for %s/%s...\n' "$release_version" "$release_os" "$release_arch"
if [ -n "$release_from_dir" ]; then
  cp "$release_from_dir/checksums.txt" "$release_work/checksums.txt"
else
  download --output "$release_work/checksums.txt" "$release_url/checksums.txt"
fi
release_expected=$(awk -v asset="$release_asset" '$2 == asset { count++; hash=$1 } END { if (count != 1) exit 1; print hash }' "$release_work/checksums.txt") || fail 'Checksum inventory is missing the archive or contains duplicates'
[[ "$release_expected" =~ ^[a-f0-9]{64}$ ]] || fail 'Invalid archive checksum'
if [ -n "$release_from_dir" ]; then
  cp "$release_from_dir/$release_asset" "$release_work/archive.tar.gz"
else
  download --output "$release_work/archive.tar.gz" "$release_url/$release_asset"
fi
[ "$(digest "$release_work/archive.tar.gz")" = "$release_expected" ] || fail 'Downloaded archive checksum does not match'

# Exactly these regular files are allowed. In particular, reject links, duplicate
# entries, directory records, absolute paths and traversal before extracting.
cat > "$release_work/expected-files" <<'FILES'
LICENSE
README.md
dependencies/Cargo.lock
dependencies/Cargo.toml
dependencies/src/main.rs
l
setup-release.sh
toolchain-version
FILES
tar -tzf "$release_work/archive.tar.gz" > "$release_work/archive-files"
sort "$release_work/archive-files" > "$release_work/sorted-files"
cmp -s "$release_work/expected-files" "$release_work/sorted-files" || fail 'Archive has missing, duplicate, or unexpected paths'
tar -tvzf "$release_work/archive.tar.gz" > "$release_work/archive-types"
awk 'substr($0,1,1) != "-" { exit 1 } END { if (NR != 8) exit 1 }' "$release_work/archive-types" || fail 'Archive contains non-regular files'
mkdir "$release_work/package"
tar -xzf "$release_work/archive.tar.gz" -C "$release_work/package"
chmod 755 "$release_work/package/l"
[ "$("$release_work/package/l" --version)" = "$release_version" ] || fail 'Binary version does not match its release'

if [ "$release_setup_mode" != binary-only ]; then
  if [ "$release_setup_mode" = with-toolchain ]; then
    bash "$release_work/package/setup-release.sh" --with-toolchain
  else
    bash "$release_work/package/setup-release.sh"
  fi
fi

mkdir -p -- "$release_prefix"
release_prefix=$(cd -- "$release_prefix" && pwd)
[ ! -L "$release_prefix/l" ] || fail 'Installation destination l is a symbolic link; choose another --prefix'
[ ! -e "$release_prefix/l" ] || [ -f "$release_prefix/l" ] || fail 'Installation destination l is not a regular file'
release_replacement=$(mktemp "$release_prefix/.l-install.XXXXXXXX")
cp "$release_work/package/l" "$release_replacement"
chmod 755 "$release_replacement"
mv -f -- "$release_replacement" "$release_prefix/l"
release_replacement=
printf 'Installed L %s at %s/l\n' "$release_version" "$release_prefix"
printf 'Add this directory to PATH in your shell profile: %s\n' "$release_prefix"
printf 'Then run: l --version\n'
if [ "$release_setup_mode" = binary-only ]; then
  printf 'Binary-only installation: l build still requires Rust and prepared locked dependencies.\n'
else
  printf 'Keep the rustup Cargo proxy on PATH: %s/bin\n' "${CARGO_HOME:-$HOME/.cargo}"
fi
