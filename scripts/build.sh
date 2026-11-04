#!/usr/bin/env sh
set -eu
bun install --frozen-lockfile
bun run typecheck
bun run build
mkdir -p bin
version="$(git rev-parse HEAD)"
go build -trimpath -ldflags "-s -w -X github.com/AlexGalhardo/Scooby/internal/trading.BuildVersion=$version" -o bin/app ./cmd/app
go build -trimpath -ldflags "-s -w -X github.com/AlexGalhardo/Scooby/internal/trading.BuildVersion=$version" -o bin/engine ./cmd/engine
