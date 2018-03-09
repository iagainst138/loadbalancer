#!/bin/bash
 set -e

#for i in {7000..7010}; do
#    python -m SimpleHTTPServer $i &
#done
#
#sleep 1
#read -p 'hit return to kill: '
#
#for pid in $(jobs -p); do
#    kill -9 $pid
#done

#wait

source tools/lib.sh

go build -i -o backends src/lb/backend/backend.go
./backends $@
