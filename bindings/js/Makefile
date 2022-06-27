SHELL=/usr/bin/env bash
VERSION=`git describe --tags --abbrev=0 | cut -b 2-`

all: build

# called by node-gyp
compile:
	CGO_CFLAGS=-mmacosx-version-min=10.13 CGO_LDFLAGS=-mmacosx-version-min=10.13 GOARCH=amd64 go build -buildmode=c-archive -o minify.a minify.go

build:
	go get -u all
	sed -i "s/\"version\": \".*\"/\"version\": \"${VERSION}\"/" package.json
	node-gyp configure
	prebuildify --napi --strip

publish:
	npm publish --access public

clean:
	rm -rf minify.a
	rm -rf build
	rm -rf prebuilds

test:
	node test.js

.PHONY: all compile build publish test clean

