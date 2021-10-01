#
# Licensed Materials - Property of IBM
#
# (c) Copyright IBM Corp. 2021.
#

all: bin/copyright bin/copyright-amd64

bin/copyright : ./*.go
	CGO_ENABLED=0 go build -o bin/copyright .

bin/copyright-amd64 : ./*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/copyright-amd64 .

clean:
	rm -rf bin
