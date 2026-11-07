#!/usr/bin/env sh
#
# The repository quality gate. Reports, never rewrites (except --format).
#
#   sh scripts/check.sh           gofmt + vet + test + lint
#   sh scripts/check.sh --no-lint skip the lint pass
#   sh scripts/check.sh --format  rewrite Go sources in place
#
# The root go.work lists every module (core, apps/cli, ...). Each Go step runs
# once per workspace module, from that module's directory.
# The lint pass is skipped rather than fatal when golangci-lint is missing, so
# a contributor without the tool can still check their work. CI installs a
# pinned golangci-lint before this script runs, so the pass always runs there.

set -eu

mode="all"
case "${1:-}" in
--no-lint) mode="nolint" ;;
--format) mode="format" ;;
"") ;;
*)
  echo "usage: $0 [--no-lint|--format]" >&2
  exit 2
  ;;
esac

root="$(git rev-parse --show-toplevel)"
cd "$root"

# Resolve the toolchain from GOROOT when it is set. mise exports GOROOT for
# the version this repo pins while leaving an unrelated `go` earlier on PATH,
# and a go binary driving a different GOROOT fails with
# `compile: version "X" does not match go tool version "Y"`.
#
# golangci-lint needs the same treatment for a different reason: it type-checks
# with whatever `go` it finds on PATH rather than with $GO, so a newer one
# there loads sources the linter's own go/types cannot parse and the pass dies
# on `file requires newer Go version goX (application built with goY)`. That
# reads as a broken gate rather than as a toolchain nobody lined up, which is
# how it ends up bypassed with --no-verify.
if [ -n "${GOROOT:-}" ] && [ -x "$GOROOT/bin/go" ]; then GO="$GOROOT/bin/go"; else GO=go; fi
if [ -n "${GOROOT:-}" ] && [ -x "$GOROOT/bin/gofmt" ]; then GOFMT="$GOROOT/bin/gofmt"; else GOFMT=gofmt; fi
if [ -n "${GOROOT:-}" ] && [ -x "$GOROOT/bin/go" ]; then LINT_PATH="$GOROOT/bin:$PATH"; else LINT_PATH="$PATH"; fi

# Every Go invocation here carries -tags goolm, and so does every release
# (.goreleaser.yml builds.flags). maunium.net/go/mautrix, which the chat
# surface imports, defaults to libolm: an all-cgo package that CGO_ENABLED=0
# cannot compile. Without the tag the build, the vet pass and the test run all
# fail on an import error that names a package nobody in this repository
# wrote, which reads as a broken tree rather than as a missing flag.
TAGS="-tags goolm"
export GOFLAGS="-tags=goolm"

if ! command -v "$GO" >/dev/null 2>&1; then
  echo "check: no usable go ('$GO')" >&2
  exit 1
fi

# This repository's own Go files: tracked plus untracked-but-not-ignored,
# minus anything deleted in the working tree but not yet staged. `git
# ls-files -c` still lists an index entry whose file is gone, and gofmt aborts
# on the missing path — xargs reports that as a non-zero exit and `set -e`
# turns it into a failed gate naming a file nobody can open.
#
# The filter batches through `sh -c 'for f do ...'` rather than `read -d`,
# because this script runs under /bin/sh and dash's read has no -d.
repo_go_files() {
  git ls-files -co --exclude-standard -z -- '*.go' |
    xargs -0 -n64 sh -c 'for f do [ -e "$f" ] || continue; printf "%s\0" "$f"; done' _
}

if [ "$mode" = "format" ]; then
  repo_go_files | xargs -0 "$GOFMT" -w
  echo "==> formatted"
  exit 0
fi

status=0

modules="$("$GO" list -m -f '{{.Dir}}')"

each_module() {
  for dir in $modules; do
    (cd "$dir" && "$@") || return 1
  done
}

# gofmt is pointed at this repository's own sources rather than at the tree,
# because the tree is not only this repository. A git worktree checked out
# below the root — which is where agent tooling puts one, under .claude/ — is
# walked by `gofmt -l .` like any other directory, so somebody else's
# half-written file failed this gate and the pre-push hook with it, naming a
# path the person pushing had never opened.
#
# Tracked and untracked-but-not-ignored, so a source file written a minute ago
# is still checked and anything .gitignore already excludes is not. That is the
# same set every other tool here means by "this repository's files".
#
# The handoff is null-delimited (-z feeding xargs -0), not whitespace-delimited:
# git passes spaces through unquoted, so xargs would split a filename at its
# space and gofmt would format half a path. It also means an empty list runs
# gofmt on nothing rather than once with stdin attached, which GNU xargs would
# do where the BSD one on macOS does not.
echo "==> version"
release="$(sh scripts/version.sh)"
cli="$(sed -n 's/^var version = "\(.*\)"$/\1/p' apps/cli/main.go)"
if [ "$release" != "$cli" ]; then
  echo "version: apps/cli/main.go says $cli but apps/client/package.json says $release"
  status=1
fi

echo "==> gofmt"
unformatted="$(repo_go_files | xargs -0 "$GOFMT" -l)"
if [ -n "$unformatted" ]; then
  echo "gofmt: the following files are not formatted (run 'sh scripts/check.sh --format'):"
  echo "$unformatted"
  status=1
fi

echo "==> go build"
each_module "$GO" build $TAGS ./... || status=1

echo "==> go vet"
each_module "$GO" vet $TAGS ./... || status=1

# -race, not as a nicety: bubbletea updates arrive from its own goroutines,
# which is exactly the shape the detector exists for.
echo "==> go test"
each_module "$GO" test $TAGS -race ./... || status=1

if [ "$mode" != "nolint" ]; then
  echo "==> golangci-lint"
  if golangci-lint version >/dev/null 2>&1; then
    PATH="$LINT_PATH" each_module golangci-lint run ./... || status=1
  else
    echo "check: no usable 'golangci-lint', skipping the lint pass (CI installs it and runs it)" >&2
  fi
fi

if [ "$status" -ne 0 ]; then
  echo "check failed"
  exit "$status"
fi

echo "==> ok"
