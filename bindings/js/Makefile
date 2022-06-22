target=./build/Release/minify.node

all: $(target)

build/:
	go get -u all
	go build -buildmode=c-archive -o minify.a minify.go

	version=$(cat go.mod | grep github.com/tdewolff/minify/v2 | cut -d ' ' -f 2 | cut -b 2-)
	sed -i "s/0.0.0/$version/" package.json

	node-gyp configure

$(target): build/ *.cc *.h *.a
	node-gyp build --target_arch=ia32
	npm run-script package --target_arch=ia32

publish:
	npm publish --access public

clean:
	rm -rf build

test:
	npm test

.PHONY: test clean

