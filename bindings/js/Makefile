target=./build/Release/minify.node

all: $(target)

build/:
	go get -u all
	go build -buildmode=c-archive -o minify.a minify.go

	version=$(cat go.mod | grep github.com/tdewolff/minify/v2 | cut -d ' ' -f 2 | cut -b 2-)
	sed -i "s/0.0.0/$version/" package.json

	node-gyp configure

$(target): build/ *.c *.h *.a
	node-gyp build

publish:
	npm publish --access public

clean:
	rm -rf build

test:
	node test.js

.PHONY: test clean

