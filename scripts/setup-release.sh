#!/usr/bin/env bash
# Runs from an extracted release archive; no compiler repository or Go is needed.
set -euo pipefail
export LC_ALL=C

fail() { printf 'L prerequisite setup failed: %s\n' "$*" >&2; exit 1; }
release_allow_install=false
case "${1:-}" in
  --with-toolchain) release_allow_install=true; shift ;;
  --help|-h)
    printf 'Usage: bash setup-release.sh [--with-toolchain]\nPrepare locked dependencies; ask before installing missing Rust tools.\n'
    exit 0 ;;
esac
[ "$#" -eq 0 ] || fail 'Unknown argument; use --help'
release_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
for release_file in toolchain-version dependencies/Cargo.toml dependencies/Cargo.lock dependencies/src/main.rs; do
  [ -f "$release_root/$release_file" ] && [ ! -L "$release_root/$release_file" ] || fail "Missing or unsafe release setup file: $release_file"
done
release_toolchain=$(cat "$release_root/toolchain-version")
[[ "$release_toolchain" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail 'Invalid pinned Rust version'
case "$(uname -s)" in
  Darwin) release_tools_hint='Install Xcode Command Line Tools with xcode-select --install, then retry.' ;;
  Linux) release_tools_hint="Install your distribution's C compiler and linker development tools, then retry." ;;
  *) fail 'Supported systems are macOS and Linux' ;;
esac
command -v cc >/dev/null 2>&1 && cc --version >/dev/null 2>&1 || fail "$release_tools_hint"
release_cargo_home=${CARGO_HOME:-${HOME:?HOME must identify your user directory}/.cargo}
export PATH="$release_cargo_home/bin:$PATH"
release_needs_install=false
if ! command -v rustup >/dev/null 2>&1 || ! RUSTUP_AUTO_INSTALL=0 rustup run "$release_toolchain" rustc --version >/dev/null 2>&1; then release_needs_install=true; fi
if [ "$release_needs_install" = true ] && [ "$release_allow_install" != true ]; then
  [ -t 0 ] || fail 'Rust tools are missing. Rerun with --with-toolchain to allow installation, or install the pinned Rust toolchain yourself.'
  printf 'Install missing rustup/Rust %s tools for this user? [y/N] ' "$release_toolchain"
  read -r release_reply
  case "$release_reply" in y|Y|yes|YES) release_allow_install=true ;; *) fail 'Rust installation declined; no compiler binary was replaced' ;; esac
fi

release_work=$(mktemp -d "${TMPDIR:-/tmp}/l-setup.XXXXXXXX")
cleanup() { rm -rf -- "$release_work"; }
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
if ! command -v rustup >/dev/null 2>&1; then
  [ "$release_allow_install" = true ] || fail 'rustup installation is not authorized'
  command -v curl >/dev/null 2>&1 || fail 'curl is required to install rustup'
  curl --proto '=https' --proto-redir '=https' --tlsv1.2 --fail --silent --show-error --location --retry 3 https://sh.rustup.rs --output "$release_work/rustup-init.sh"
  sh "$release_work/rustup-init.sh" -y --no-modify-path --profile minimal --default-toolchain none
fi
if ! RUSTUP_AUTO_INSTALL=0 rustup run "$release_toolchain" rustc --version >/dev/null 2>&1; then
  [ "$release_allow_install" = true ] || fail 'Rust toolchain installation is not authorized'
  rustup toolchain install "$release_toolchain" --profile minimal --no-self-update
fi
command -v cargo >/dev/null 2>&1 || fail 'rustup Cargo proxy is missing from PATH'
RUSTUP_AUTO_INSTALL=0 cargo "+$release_toolchain" --version >/dev/null || fail 'Cargo must be the rustup proxy supporting the pinned toolchain'

# Fetch from a disposable seed with the reviewed release lock. Never generate or
# rewrite a lock, copy CI caches, or change Cargo's offline compiler contract.
mkdir -p "$release_work/dependencies/src"
cp "$release_root/dependencies/Cargo.toml" "$release_work/dependencies/Cargo.toml"
cp "$release_root/dependencies/Cargo.lock" "$release_work/dependencies/Cargo.lock"
cp "$release_root/dependencies/src/main.rs" "$release_work/dependencies/src/main.rs"
printf 'Preparing locked Rust %s dependencies...\n' "$release_toolchain"
CARGO_NET_OFFLINE=false RUSTUP_AUTO_INSTALL=0 cargo "+$release_toolchain" fetch --locked --target host-tuple --manifest-path "$release_work/dependencies/Cargo.toml"
cmp -s "$release_root/dependencies/Cargo.lock" "$release_work/dependencies/Cargo.lock" || fail 'Dependency fetching changed the retained lock'
CARGO_NET_OFFLINE=true RUSTUP_AUTO_INSTALL=0 cargo "+$release_toolchain" fetch --frozen --target host-tuple --manifest-path "$release_work/dependencies/Cargo.toml"
cmp -s "$release_root/dependencies/Cargo.lock" "$release_work/dependencies/Cargo.lock" || fail 'Offline verification changed the retained lock'
printf 'Native prerequisites are prepared. Keep %s/bin on PATH.\n' "$release_cargo_home"
