SHELL=/usr/bin/env bash
VERSION=`git describe --tags --abbrev=0`

all: build

build: *.go MANIFEST.in
	go get github.com/tdewolff/minify/v2@${VERSION}
	go build -buildmode=c-shared -o minify.so
	python -m build --sdist

publish: clean build
	twine upload dist/*

clean:
	rm -rf dist
	rm -rf tdewolff_minify.egg-info

.PHONY: all build publish clean

