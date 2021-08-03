#!/bin/sh
set -e

echo "building normal version to..."
go build -o covid-cert
echo "building main.wasm..."
GOOS=js GOARCH=wasm go build -o main.wasm
echo "building webserver..."
go build -o http-server cmd/server/main.go
