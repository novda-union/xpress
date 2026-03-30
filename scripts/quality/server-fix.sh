#!/usr/bin/env sh
set -eu

echo "server: fmt"
find server -type f -name '*.go' -print | xargs gofmt -w
