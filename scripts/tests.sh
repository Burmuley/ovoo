#!/usr/bin/env bash

set -e

echo "Running unit-tests"

packages=(
    "internal/repositories/drivers/gorm"
    "internal/services"
    "internal/entities"
)

for pkg in "${packages[@]}"; do
    echo "Testing $pkg"
    go test "$PWD/$pkg"
done
