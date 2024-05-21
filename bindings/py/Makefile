SHELL=/usr/bin/env bash
VERSION=`git describe --tags --abbrev=0 | cut -b 2-`

all: build

compile:
	# called from workflow
	sed -i.bak -e "s/{VERSION}/${VERSION}/" setup.py
	go build -buildmode=c-shared -o src/minify/minify.so

build: compile
	python -m build --sdist

publish: clean build
	twine upload dist/*

clean:
	rm -rf dist
	rm -rf tdewolff_minify.egg-info
	rm -rf go.*

.PHONY: all compile build publish clean

