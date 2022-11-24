#
# Licensed Materials - Property of IBM
#
# (c) Copyright IBM Corp. 2021.
#

all: bin/copyright bin/copyright-amd64 bin/copyright-arm64

bin/copyright : ./Makefile ./*.go
	CGO_ENABLED=0 go build -o bin/copyright .

bin/copyright-amd64 : ./Makefile ./*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/copyright-amd64 .

bin/copyright-arm64 : ./Makefile ./*.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/copyright-arm64 .


clean:
	rm -rf bin
