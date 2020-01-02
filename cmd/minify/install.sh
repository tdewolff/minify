#!/bin/bash

go install -ldflags "-X 'main.Version=devel' -X 'main.Commit=$(git rev-list -1 HEAD)' -X 'main.Date=$(date)'"

source minify_bash_tab_completion
