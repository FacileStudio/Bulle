#!/bin/sh
# Prints the release version, read from apps/client/package.json, the file the
# release commit bumps, with the leading v that /version reports.
set -eu
here="$(cd "$(dirname "$0")/.." && pwd)"
version="$(sed -n 's/^[[:space:]]*"version":[[:space:]]*"\([^"]*\)".*/\1/p' "$here/apps/client/package.json" | head -n 1)"
[ -n "$version" ] || { echo "version.sh: no version in apps/client/package.json" >&2; exit 1; }
echo "v$version"
