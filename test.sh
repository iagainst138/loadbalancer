#!/bin/bash

set -eo pipefail

source tools/lib.sh

SET_GOPATH

mkdir -p coverage
REPORT=coverage/coverage.html

function GEN_REPORT() {
    TOTAL_COVERAGE=$(grep '^coverage: .*statements$' | awk '{print $2}')
    # generate the coverage html file
    go tool cover -html=coverage/cover.out -o $REPORT
    # add the overall coverage to the report
    sed -i "s#<span class=\"cov8\">covered</span>.*#<span class=\"cov8\">covered</span> <span style=\"color:white;\">Overall Coverage: $TOTAL_COVERAGE</span>#g" $REPORT
}

go test lb  -v -cover -coverprofile=coverage/cover.out $@ | tee >(GEN_REPORT)
