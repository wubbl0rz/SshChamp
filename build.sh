#/bin/bash

cd "$(dirname "$0")"

rm -rf ./go_build

cd ./app

CGO_ENABLED=0 GOOS=linux go build -o ../go_build/app .

