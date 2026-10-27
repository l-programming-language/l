#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 2 || $(uname -s) != Darwin ]]; then
  printf 'Usage on macOS: sign-macos.sh BINARY RECEIPT.json\n' >&2
  exit 1
fi
binary=$1
receipt=$2
: "${SIGNING_IDENTITY:?Developer ID Application identity is required}"
: "${APPLE_ID:?Apple notarization account is required}"
: "${APPLE_TEAM_ID:?Apple developer team is required}"
: "${APPLE_APP_PASSWORD:?Apple app-specific password is required}"
[[ -f "$binary" && ! -L "$binary" && ! -e "$receipt" ]] || { printf 'Expected a regular binary and new receipt path.\n' >&2; exit 1; }
signing_temp=$(mktemp -d "${TMPDIR:-/tmp}/l-notarize.XXXXXX")
trap 'rm -rf "$signing_temp"' EXIT
codesign --force --timestamp --options runtime --sign "$SIGNING_IDENTITY" "$binary"
codesign --verify --strict --verbose=2 "$binary"
ditto -c -k --keepParent "$binary" "$signing_temp/l.zip"
xcrun notarytool submit "$signing_temp/l.zip" --apple-id "$APPLE_ID" --team-id "$APPLE_TEAM_ID" --password "$APPLE_APP_PASSWORD" --wait --output-format json > "$signing_temp/notary.json"
node --input-type=module - "$binary" "$signing_temp/notary.json" "$receipt" <<'NODE'
import fs from 'node:fs';
import { createHash } from 'node:crypto';
const [binary, result, receipt] = process.argv.slice(2);
const response = JSON.parse(fs.readFileSync(result));
if (response.status !== 'Accepted' || !response.id) throw new Error('Apple did not accept notarization');
const value = { status: response.status, notarizationId: response.id, identity: process.env.SIGNING_IDENTITY, binarySha256: createHash('sha256').update(fs.readFileSync(binary)).digest('hex') };
fs.writeFileSync(receipt, JSON.stringify(value, null, 2) + '\n', { flag: 'wx' });
NODE
# Standalone Mach-O tools cannot carry a stapled ticket. Gatekeeper retrieves the
# notarization ticket for this signed executable; the final tar preserves its bytes.
