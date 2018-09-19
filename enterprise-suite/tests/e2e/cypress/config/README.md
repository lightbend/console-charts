## Code flow for env specific config file

1. in package.json, pass confie file setting to cypress, such as ```cypress run --env configFile=minikube```
2. in cypress/plugin/index.js,  based on config.env.configFile set in step 1, decide read which config in cypress/config/
   For example, minikube.json or local.json (default), merge it with existing config
3. in cypress/support/environment.ts , use Cypress.env().configFile to decide which core config file to use (For example, use src/environments/environment.dev-release)


## reference
https://docs.cypress.io/api/plugins/configuration-api.html
https://docs.cypress.io/guides/guides/environment-variables.html
