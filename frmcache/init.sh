#!/bin/bash

test ! -e /usr/local/bin/app && /usr/local/go/bin/go build -o /usr/local/bin/app ./...

/usr/local/bin/app -hostname $FRM_HOST -port $FRM_PORT &
PID="$!"

trap 'kill $PID; exit 0' EXIT INT TERM
wait
