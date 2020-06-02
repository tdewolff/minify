module github.com/tdewolff/fuzz/minify/css-token

go 1.13

replace github.com/tdewolff/minify/v2 => ../../../minify

require (
	github.com/dvyukov/go-fuzz v0.0.0-20191022152526-8cb203812681 // indirect
	github.com/tdewolff/minify/v2 v2.5.2
	github.com/tdewolff/parse/v2 v2.4.2
)
