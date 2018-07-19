#!/usr/bin/env bash

function yq() {
    docker run -v ${PWD}:/workdir mikefarah/yq yq $@
}
