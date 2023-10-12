#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#

all: tests bin/copyright bin/copyright-amd64

source-code : \
	./cmd/githubapp-copyright/*.go \
	./pkg/checks/*.go \
	./pkg/checkTypes \
	./pkg/fileCheckers \
	./pkg/embedded

bin/copyright : source-code
	CGO_ENABLED=0 go build -o bin/copyright ./cmd/githubapp-copyright

bin/copyright-amd64 : source-code
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/copyright-amd64 ./cmd/githubapp-copyright

tests: build/coverage.txt build/coverage.html

build/coverage.out : source-code
	mkdir -p build
	go test -v -cover -coverprofile=build/coverage.out  -coverpkg ./pkg/checks,./pkg/checkTypes,./pkg/fileCheckers ./pkg/...

build/coverage.html : build/coverage.out
	go tool cover -html=build/coverage.out -o build/coverage.html

build/coverage.txt : build/coverage.out
	go tool cover -func=build/coverage.out > build/coverage.txt ;
	cat build/coverage.txt

clean:
	rm -fr bin/*
	rm -fr build/*



