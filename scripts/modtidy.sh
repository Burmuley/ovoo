#!/usr/bin/env bash

set -e

go mod tidy
pushd ./internal/tools && go mod tidy && popd
