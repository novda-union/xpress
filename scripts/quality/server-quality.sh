#!/usr/bin/env sh
set -eu

export GOCACHE="${GOCACHE:-/tmp/xpressgo-go-cache}"
export GOLANGCI_LINT_CACHE="${GOLANGCI_LINT_CACHE:-/tmp/xpressgo-golangci-lint-cache}"

GOLANGCI_LINT_BIN="${GOLANGCI_LINT_BIN:-}"
if [ -z "$GOLANGCI_LINT_BIN" ] && command -v golangci-lint >/dev/null 2>&1; then
  GOLANGCI_LINT_BIN="$(command -v golangci-lint)"
fi
if [ -z "$GOLANGCI_LINT_BIN" ]; then
  GOPATH_BIN="$(go env GOPATH 2>/dev/null)/bin/golangci-lint"
  if [ -x "$GOPATH_BIN" ]; then
    GOLANGCI_LINT_BIN="$GOPATH_BIN"
  fi
fi

if [ -z "$GOLANGCI_LINT_BIN" ]; then
  echo "server: golangci-lint is required but not installed"
  exit 1
fi

echo "server: fmt-check"
unformatted="$(find server -type f -name '*.go' -print | xargs gofmt -l || true)"
if [ -n "$unformatted" ]; then
  printf '%s\n' "$unformatted"
  echo "server: gofmt check failed"
  exit 1
fi

echo "server: vet"
(cd server && go vet ./...)

echo "server: lint"
(cd server && "$GOLANGCI_LINT_BIN" run ./...)

echo "server: test"
(cd server && go test ./...)
