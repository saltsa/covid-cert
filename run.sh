#!/bin/sh
set -e

echo "building main.wasm..."
GOOS=js GOARCH=wasm go build -o main.wasm
echo "building webserver..."
go build -o http-server cmd/server/main.go
./http-server
