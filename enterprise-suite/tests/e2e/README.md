# How to run cypress in mac to test local minikube setup

Assume that es-console is running in http://192.168.99.100:30080/


If your es-console is running in different url, please modify file `cypress/config/minikube.json`

To test it with cypress

```
# use node.js v9.4.0  (NOTE: the same version as es-console-spa to avoid install two different versions)
nvm install v9.4.0
nvm use v9.4.0

# install npm package
npm install

# set up demo app
npm run e2e:demo-app-setup

# run the gui mode with skipKnownError
npx cypress open --env configFile=minikube,skipKnownError=true

# (Optional) run the test mode
npx cypress run --env configFile=minikube,skipKnownError=true

```
