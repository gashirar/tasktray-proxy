#!/bin/sh

if [ $# != 1 ]; then
	echo "Usage: $0 [binary name]"
	exit 0
fi

GOOS=linux    GOARCH=amd64 go build -o ./bin/$1 cmd/main.go
GOOS=windows  GOARCH=amd64 go build -o ./bin/$1.exe -ldflags -H=windowsgui cmd/main.go
# GOOS=darwin GOARCH=amd64 go build cmd/main.go -o ./bin/darwin64/$1
