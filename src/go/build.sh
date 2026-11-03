#!/usr/bin/env bash
# Bygger typetrainer och installerar det i ~/.local/bin
set -euo pipefail

cd "$(dirname "$0")"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date +"%Y-%m-%dT%H:%M")

go build -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" -o typetrainer ./cmd/cli
go build -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" -o "$HOME/.local/bin/typetrainer" ./cmd/cli

echo "Byggd ${VERSION} (${BUILD_TIME}) -> $HOME/.local/bin/typetrainer"
