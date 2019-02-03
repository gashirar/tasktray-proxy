#!/bin/sh

if [ $# != 1 ]; then
	echo "Usage: $0 [binary name]"
	exit 0
fi

GOOS=linux    GOARCH=amd64 go build -o ./bin/$1
GOOS=windows  GOARCH=amd64 go build -o ./bin/$1.exe -ldflags -H=windowsgui
# GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin64/$1
