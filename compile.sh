#!/bin/sh
set -e

echo "building native covid-cert application.."
go build -o covid-cert
echo "building webassembly main.wasm..."
GOOS=js GOARCH=wasm go build -o main.wasm
echo "building webserver..."
go build -o http-server cmd/server/main.go
