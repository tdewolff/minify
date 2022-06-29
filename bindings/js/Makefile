SHELL=/usr/bin/env bash
VERSION=`git describe --tags --abbrev=0 | cut -b 2-`

# node-gyp will run c++ with -mmacosx-version-min=10.13, set golang to match this
ifeq ($(shell uname -s),Darwin)
export CGO_CFLAGS += -mmacosx-version-min=10.13
export CGO_LDFLAGS += -mmacosx-version-min=10.13
endif

all: build

# called by node-gyp
compile:
	go build -buildmode=c-archive -o minify.a minify.go

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

