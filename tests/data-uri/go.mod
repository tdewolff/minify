module github.com/tdewolff/minify/tests/data-uri

go 1.13

replace github.com/tdewolff/parse/v2 => ../../../parse

replace github.com/tdewolff/minify/v2 => ../../../minify

require (
	github.com/dvyukov/go-fuzz v0.0.0-20200318091601-be3528f3a813 // indirect
	github.com/tdewolff/minify/v2 v2.7.6
	github.com/tdewolff/parse/v2 v2.4.3
)
