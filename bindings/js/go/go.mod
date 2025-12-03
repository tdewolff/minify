module github.com/tdewolff/minify/v2/bindings/js/go

go 1.24.0

toolchain go1.24.1

require (
	github.com/tdewolff/minify/v2 v2.24.7
	github.com/tdewolff/parse/v2 v2.8.5
)

replace github.com/tdewolff/minify/v2 => ../../..
