SHELL=/usr/bin/env bash
HASH=`git log -1 --format="%H"`
VERSION=`git describe --tags --abbrev=0 | cut -b 2-`

# needed explicitly for Windows
export CGO_ENABLED=1

# node-gyp will run c++ with -mmacosx-version-min=10.13, set golang to match this
ifeq ($(shell uname -s),Darwin)
export CGO_CFLAGS += -mmacosx-version-min=10.13
export CGO_LDFLAGS += -mmacosx-version-min=10.13
endif

all: build

# called by node-gyp for build-from-source
compile:
	go mod init github.com/tdewolff/minify/bindings/js
	go get -u github.com/tdewolff/parse/v2@master
	go get -u github.com/tdewolff/minify/v2@master
	go mod tidy
	go build -buildmode=c-archive -o minify.a minify.go

build:
	node-gyp configure
	prebuildify --napi --strip
	rm -rf go.*

publish: build
	sed -i.bak -e "s/{VERSION}/${VERSION}/" package.json
	npm publish --access=public
	mv package.json.bak package.json

clean:
	rm -rf minify.a
	rm -rf build
	rm -rf prebuilds
	rm -rf go.*

test: build
	npm test

.PHONY: all compile build publish clean test

