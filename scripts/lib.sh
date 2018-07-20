#!/usr/bin/env bash

if ! command -v yq > /dev/null; then
    >&2 echo "Please install yq:"
    >&2 echo "* brew install yq" 
    >&2 echo "* go get github.com/mikefarah/yq"
    >&2 echo "See https://github.com/mikefarah/yq for more options"
    echo "yq_missing"
    exit 1
fi
