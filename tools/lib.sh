#!/bin/bash

function SET_GOPATH() {
    if which cygpath &> /dev/null;then # on windows
        export GOPATH="$(cygpath -w $(dirname $(readlink -f $0)));$(go env GOPATH)"
    else
        export GOPATH=$(dirname $(readlink -f $0)):$(go env GOPATH)
    fi
}
