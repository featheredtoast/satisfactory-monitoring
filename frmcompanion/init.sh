#!/bin/bash

test ! -e Companion && /usr/bin/git clone --depth 1 -b multiarch https://github.com/featheredtoast/FicsitRemoteMonitoringCompanion.git .;
cd Companion;
mkdir -p ./bin/map
test ! -e bin/companion && /usr/local/go/bin/go build main.go -o bin/companion
apt update
apt install -y nodejs npm
cd ../map
npm install
npm run compile
cp -R index.html map-16k.png vendor/ img/ js/ ../Companion/bin/map
cd ../Companion/bin
./companion -hostname fakeserver -port 8081
