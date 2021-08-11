#!/bin/sh
set -e

echo "building native covid-cert application.."
go build -o covid-cert ./cmd/app
echo "building webassembly main.wasm..."
GOOS=js GOARCH=wasm go build -o static/main.wasm ./cmd/app
echo "building webserver..."
go build -o http-server ./cmd/server
