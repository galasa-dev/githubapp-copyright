#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#


all: tests bin/copyright bin/copyright-amd64

source-code : ./cmd/githubapp-copyright/*.go ./pkg/checks/*.go

bin/copyright : source-code
	CGO_ENABLED=0 go build -o bin/copyright ./cmd/githubapp-copyright

bin/copyright-amd64 : source-code
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/copyright-amd64 ./cmd/githubapp-copyright

tests: build/coverage.txt build/coverage.html

build/coverage.out : source-code
	mkdir -p build
	# Compile the tests into a binary executable "checks.test"
	go test -c -v -cover  -coverpkg ./pkg/checks ./pkg/...
	# Keep it tidy by moving it out the way so it never gets checked-in.
	mv checks.test build
	# Each line in the makefile executes with a different environment.
	# So cd'ing to a folder has no effect unless you do something immediately on the
	# same line... 
	cd build ; ./checks.test -test.coverprofile coverage.out

build/coverage.html : build/coverage.out
	go tool cover -html=build/coverage.out -o build/coverage.html

build/coverage.txt : build/coverage.out
	go tool cover -func=build/coverage.out > build/coverage.txt ;
	cat build/coverage.txt

clean:
	rm -fr bin/*
	rm -fr build/*



