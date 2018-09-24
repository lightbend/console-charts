// ***********************************************************
// This example plugins/index.js can be used to load plugins
//
// You can change the location of this file or turn off loading
// the plugins file with the 'pluginsFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/plugins-guide
// ***********************************************************

// This function is called when a project is opened or re-opened (e.g. due to
// the project's config changing)
const webpack = require('@cypress/webpack-preprocessor')
const cucumber = require('cypress-cucumber-preprocessor').default
const fs = require('fs-extra')
const path = require('path')

function getConfigurationByFile (env) {
  const pathToConfigFile = path.resolve('cypress', 'config', `${env}.json`)
  return fs.readJson(pathToConfigFile)
}

module.exports = (on, config) => {
  // `on` is used to hook into various events Cypress emits
  // `config` is the resolved Cypress config
  const options = {
    webpackOptions: require('../../webpack.cypress'),
  }

  on('file:preprocessor', (file) => {
    if (file.filePath.match(/step_definitions/g)) {
      return cucumber()(file)
    } else {
      return webpack(options)(file)
    }
  })

  // merge default config and custom config file (like local.json or minikube.json)
  config.env.configFile = config.env.configFile || 'local'
  const env = config.env.configFile
  console.log('env: ', env)
  return getConfigurationByFile(env).then((envConfig) => {
    const finalEnvSetting = Object.assign(config.env, envConfig.env)
    const finalConfig = Object.assign(config, envConfig, {env: finalEnvSetting})
    console.log('final config: ', finalConfig)
    return finalConfig
  })

}
