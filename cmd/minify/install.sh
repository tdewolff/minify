#!/bin/bash

go install -ldflags "-s -w -X 'main.Version=devel' -X 'main.Commit=$(git rev-list -1 HEAD)'"

source minify_bash_tab_completion
