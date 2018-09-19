#!/usr/bin/env bash

set -exu

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# setup
echo "Installing ES from helm charts, $( basename $( pwd ) )"

make build
make install-local

# run tests
echo "Running tests"
cd $script_dir/../tests/e2e

# install nvm
wget -qO- https://raw.githubusercontent.com/creationix/nvm/v0.33.11/install.sh | bash

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion
nvm version

nvm install v9.4.0
nvm use v9.4.0

# run the e2e test
npm install
npm run e2e:demo-app-setup
npm run e2e:travis-prs
