all: build

build: *.go *.c *.h
	go get -u all
	go build -buildmode=c-archive -o minify.a minify.go

	version=$(cat go.mod | grep github.com/tdewolff/minify/v2 | cut -d ' ' -f 2 | cut -b 2-)
	sed -i "s/0.0.0/$version/" package.json

	node-gyp configure
	prebuildify --napi

publish:
	npm publish --access public

clean:
	rm -rf minify.a
	rm -rf build
	rm -rf prebuilds

test:
	node test.js

.PHONY: all build publish test clean

