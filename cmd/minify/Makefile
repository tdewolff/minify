SHELL=/bin/bash

install:
	go build -ldflags "-X 'main.Version=devel' -X 'main.Commit=$(shell git rev-list -1 HEAD)' -X 'main.Date=$(shell date)'" -o minify

	source minify_bash_tab_completion

clean:
	rm minify
