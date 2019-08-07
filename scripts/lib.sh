#!/usr/bin/env bash

if yqpath=$(command -v yq); then
    if command -v go; then
        go get github.com/mikefarah.yq
        yqpath=$(command -v yq)
    fi
fi

yq() {
    if [ "$yqpath" != "" ]; then
        "$yqpath" $@
    else
        echo "No yq or docker available, bailing..."
        exit 1
    fi
}
