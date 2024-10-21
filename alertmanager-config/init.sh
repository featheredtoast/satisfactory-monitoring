#!/bin/bash

test ! -e /usr/local/bin/app && /usr/local/go/bin/go build -o /usr/local/bin/app ./...

/usr/local/bin/app
