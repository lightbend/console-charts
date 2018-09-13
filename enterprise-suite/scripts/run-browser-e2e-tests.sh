# install nvm
wget -qO- https://raw.githubusercontent.com/creationix/nvm/v0.33.11/install.sh | bash

source ~/.profile
nvm version

nvm install v9.4.0
nvm use v9.4.0

git clone git@github.com:lightbend/es-console-spa.git
cd es-console-spa
npm install
npm run e2e:travis-prs
