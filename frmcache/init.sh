#!/bin/bash

go install github.com/jackc/tern@latest
tern migrate --config /usr/src/db/tern.conf --migrations /usr/src/db

test ! -e /usr/local/bin/app && /usr/local/go/bin/go build -o /usr/local/bin/app ./...

/usr/local/bin/app -hostname $FRM_HOST -port $FRM_PORT -hostnames $FRM_HOSTS &
PID="$!"

trap 'kill $PID; exit 0' EXIT INT TERM
wait
