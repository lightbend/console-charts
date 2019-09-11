#!/usr/bin/env bash

set -e

version="v0.10.0"

mkdir -p $GOPATH/src/github.com/operator-framework
cd $GOPATH/src/github.com/operator-framework
git clone --branch ${version} --depth 1 https://github.com/operator-framework/operator-sdk
cd operator-sdk
make install
