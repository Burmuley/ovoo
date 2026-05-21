#!/usr/bin/env bash

set -e

echo "Running unit-tests"

packages=(
    "internal/repositories/drivers/*"
    "internal/services/*"
    "internal/entities/*"
    "internal/cache/drivers/*"
    "internal/config"
    "internal/applications/milter"
    "internal/applications/rest"
)

for pkg in "${packages[@]}"; do
    echo "Testing $pkg"
    go test "$PWD"/$pkg
done
