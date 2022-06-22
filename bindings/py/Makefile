all: build

build: *.c *.go
	go get -u all
	go build -buildmode=c-shared -o minify.so
	python -m build --sdist

publish: clean build
	twine upload dist/*

clean:
	rm -rf dist
	rm -rf tdewolff_minify.egg-info

.PHONY: publish clean

