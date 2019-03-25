#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CGO_ENABLED=0 go test $(go list ./... | grep -v -e '/client/' -e '/samples/' -e '/apis/')
