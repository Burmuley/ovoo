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
    pushd "$WEBUI_DIR"
    npm run build && cp -Rf dist/* ../internal/applications/rest/data/webui
    popd
}


cp -Rf "$WEBUI_DIR"/dist/* "$OVOO_API_DATA_DIR"

go generate "$PWD/internal/applications/rest"
go generate "$PWD/internal/config"

if [[ "${TARGET}" == "local" ]]; then
    build_webui
    go build -o "bin/ovoo_${TARGET}" ./cmd/ovoo
elif [[ "${TARGET}" == "linux" ]]; then
    build_webui
    GOOS=linux GOARCH=amd64 go build -o bin/ovoo_linux ./cmd/ovoo
elif [[ "${TARGET}" == "webui" ]]; then
    build_webui
fi
