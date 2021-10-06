#!/bin/bash
set -e
mkdir -p builds/test
cd ui
yarn build
cd ..
rm -rf ./src/managers/server/ui
cp -r ./ui/build ./src/managers/server/ui
cd src
go build -o ../builds/test/fakeserver ./fakeserver/fakeserver.go
go build -o ../builds/test/ftg ./forgetogo/main.go
cd ..

if [ -d "./sandbox" ]; then
    rm -rf ./sandbox
fi
mkdir -p sandbox/bin
cd sandbox
cp ../builds/test/fakeserver ./bin/java
cp ../builds/test/ftg .
PATH="./bin:${PATH}" ./ftg
cd ..