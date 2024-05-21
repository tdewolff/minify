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
	go mod tidy
	NODE_VERSION=`node -e 'console.log(require("./package.json").version);'`
	go get github.com/tdewolff/minify/v2@v${NODE_VERSION}
	go build -buildmode=c-archive -o minify.a minify.go

build:
	go mod init github.com/tdewolff/minify/bindings/js
	go mod tidy
	go get github.com/tdewolff/minify/v2@${HASH}
	node-gyp configure
	prebuildify --napi --strip

publish:
	sed -i.bak -e "s/\"version\": \".*\"/\"version\": \"${VERSION}\"/" package.json
	rm package.json.bak
	npm publish --access=public

clean:
	rm -rf minify.a
	rm -rf build
	rm -rf prebuilds
	rm -rf go.*

test:
	npm test

.PHONY: all compile build publish clean test

