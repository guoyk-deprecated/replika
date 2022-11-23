#!/bin/bash

set -eu

cd $(dirname $0)

rm -rf dist
mkdir -p dist

build() {
  export GOOS=$1
  export GOARCH=$2
  export CGO_ENABLED=0
  PACKNAME="replika-${GOOS}-${GOARCH}"
  FILENAME="replika"
  go build -o "dist/${FILENAME}" .
  cd dist
  tar czvf "${PACKNAME}.tar.gz" "${FILENAME}"
  rm -f "${FILENAME}"
  cd ..
}

build linux amd64
build windows amd64
build darwin amd64
build darwin arm64

build linux arm64
build linux loong64
build windows 386
build freebsd amd64
