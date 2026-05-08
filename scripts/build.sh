#!/usr/bin/env bash

set -e

TARGET="$1"

if [[ "${TARGET}" == "" ]]; then
    TARGET="local"
fi

echo "Building for target: ${TARGET}"

WEBUI_DIR="$PWD/webui"
OVOO_API_DATA_DIR="$PWD/internal/applications/rest/data/webui"

function build_webui () {
    npm install -C "$WEBUI_DIR"
    npm run build -C "$WEBUI_DIR"
    rm -rf "${OVOO_API_DATA_DIR:?}/*"
    cp -Rf "$WEBUI_DIR"/dist/* "$OVOO_API_DATA_DIR"
}

go generate "$PWD/internal/applications/rest"

if [[ "${TARGET}" == "local" ]]; then
    build_webui
    rm -f bin/ovoo_linux
    go build -ldflags="-s -w" -o "bin/ovoo_${TARGET}" ./cmd/ovoo
elif [[ "${TARGET}" == "linux" ]]; then
    build_webui
    rm -f bin/ovoo_linux
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o bin/ovoo_linux ./cmd/ovoo
    upx bin/ovoo_linux
elif [[ "${TARGET}" == "webui" ]]; then
    build_webui
fi
