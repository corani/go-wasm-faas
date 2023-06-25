#!/bin/bash

if [[ "$(go tool dist list | grep wasip1/wasm)" == "wasip1/wasm" ]]; then
    WASIGO=go
else
    command -v gotip > /dev/null
    if [[ $? == 1 ]]; then
        echo "WASI requires Go 1.21 or later. Install Go 1.21 or use 'gotip':"
        echo "go install golang.org/dl/gotip@latest"
        echo "gotip download"
        exit 1
    else
        WASIGO=gotip
    fi
fi

echo "[INFO] building wasm module using $WASIGO"
GOOS=wasip1 GOARCH=wasm $WASIGO build -o target/goenv.wasm module/main.go
ls -lAh target/

echo "[INFO] building server"
go build -o bin/server server/main.go
ls -lAh bin/
