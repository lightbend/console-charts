#!/usr/bin/env bash

set -e

RELEASE_VERSION="v0.10.0"
curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
sudo mv operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu /usr/local/bin/operator-sdk
sudo chmod +x /usr/local/bin/operator-sdk
