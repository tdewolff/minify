#!/bin/bash

go install -ldflags "-s -w -X 'main.Version=$(git describe --tags)'" -trimpath

source minify_bash_tab_completion
