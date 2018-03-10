#!/bin/bash

set -e

source tools/lib.sh

BUILD_DIR=builds
VERSION=$(cat version.txt)
GIT_COMMIT=$(git rev-parse HEAD)
BUILD_TARGETS=${BUILD_TARGETS:-"windows linux darwin"}

function _cleanup() {
    r=$?
    if [ $r -eq 0 ];then
        echo "BUILD OK"
    else
        echo "BUILD FAILED"
    fi
    exit $r
}

trap _cleanup EXIT

rm -rf ${BUILD_DIR}

SET_GOPATH

go run src/lb/cmd/generate.go

for PLATFORM in ${BUILD_TARGETS}; do
    OUT=lb-${VERSION}
    [ ${PLATFORM} == "windows" ] && OUT=${OUT}.exe
    mkdir -p ${BUILD_DIR}/${PLATFORM}
    GOOS=$PLATFORM GOARCH=amd64 \
        go build -ldflags="-X lb.Version=${VERSION}/${GIT_COMMIT}" \
        -i -o ${BUILD_DIR}/${PLATFORM}/${OUT} src/lb/cmd/main.go
done
