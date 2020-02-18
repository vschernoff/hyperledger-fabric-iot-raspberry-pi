#!/usr/bin/env bash

echo "Cloning fabric-rest-api-go into ./.tmp folder and build frag:latest docker container"

mkdir -p .tmp

cd .tmp

git clone git@gitlab.altoros.com:intprojects/fabric-rest-api-go.git

cd fabric-rest-api-go

docker build -t frag:latest .
