#!/usr/bin/env bash

if ! command -v yq > /dev/null; then
    echo "Please install yq:"
    echo "* brew install yq"
    echo "* go get github.com/mikefarah/yq"
    echo "See https://github.com/mikefarah/yq for more options"
    exit 1
fi
