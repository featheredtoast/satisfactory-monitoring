#!/bin/bash

test ! -e Companion && /usr/bin/git clone --depth 1 -b unregister-old-metrics https://github.com/featheredtoast/FicsitRemoteMonitoringCompanion.git .
cd Companion
mkdir -p ./bin/map
test ! -e ./bin/companion && /usr/local/go/bin/go build -o bin/companion main.go
if [ -z "$INSTALL_MAP"]
then
    echo "Skipping map install. Set INSTALL_MAP var to true to enable."
else
    test ! -e /usr/bin/npm && apt update && apt install -y nodejs npm
    cd ../map
    npm install
    npm run compile
    cp -R index.html map-16k.png vendor/ img/ js/ ../Companion/bin/map
fi
cd ../Companion/bin
touch ./frmc.log
./companion &
PID="$!"
trap 'kill $PID; exit 0' EXIT INT TERM
tail -f ./frmc.log &
PID2="$!"
trap 'kill $PID; exit 0' EXIT INT TERM
wait
