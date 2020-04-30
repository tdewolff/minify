#!/bin/sh

if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
    go install -ldflags "-s -w -X 'main.Version=$(git describe --tags)'" -trimpath
else
    go install -ldflags "-s -w" -trimpath
fi

. minify_bash_tab_completion
