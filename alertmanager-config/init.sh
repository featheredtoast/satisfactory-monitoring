#!/bin/bash

test ! -e /usr/local/bin/app && /usr/local/go/bin/go build -o /usr/local/bin/app ./...

OUTPUT_PATH=/etc/alertmanager/alertmanager.yml INPUT_PATH=/usr/src/templates/alertmanager.yml.tmpl /usr/local/bin/app
