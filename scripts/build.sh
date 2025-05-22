#!/usr/bin/env bash

set -e

TARGET="$1"

if [[ "${TARGET}" == "" ]]; then
    TARGET="local"
fi

echo "Building for target: ${TARGET}"

WEBUI_DIR="$PWD/webui"
OVOO_API_DATA_DIR="$PWD/internal/applications/rest/data/webui"

pushd "$WEBUI_DIR"
npm run build
popd

cp -Rf "$WEBUI_DIR"/dist/* "$OVOO_API_DATA_DIR"

if [[ "${TARGET}" == "local" ]]; then
    go build -o "bin/ovoo_${TARGET}" ./cmd/ovoo
elif [[ "${TARGET}" == "linux" ]]; then
    GOOS=linux GOARCH=amd64 go build -o bin/ovoo_linux ./cmd/ovoo
fi
