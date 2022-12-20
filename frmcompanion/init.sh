#!/bin/bash

test ! -e Companion && /usr/bin/git clone --depth 1 -b multiarch https://github.com/featheredtoast/FicsitRemoteMonitoringCompanion.git .;
cd Companion;
/usr/local/go/bin/go run main.go -hostname fakeserver -port 8081
