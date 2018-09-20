# How to run cypress in mac to test local minikube setup

Assume that es-console is running in http://192.168.99.100:30080/

To test it with cypress

```
# use node.js v10.8.0
nvm install v10.8.0
nvm use v10.8.0

# install npm package
npm ci   # "npm ci" is equivalent to "npm install" but much faster

# set up demo app
npm run e2e:demo-app-setup

# run the gui mode with skipKnownError
npx cypress open --env configFile=minikube,skipKnownError=true

# (Optional) run the test mode
npx cypress run --env configFile=minikube,skipKnownError=true

```
