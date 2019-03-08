import { base, Env } from './environment.base';

// The file contents for the current environment will overwrite these during build.
// The build system defaults to the dev environment which uses `environment.ts`, but if you do
// `ng build --env=prod` then `environment.prod.ts` will be used instead.
// The list of which env maps to which file can be found in `.angular-cli.json`.

export const environment: Env = {
  ...base,
  production: false,
  prometheusApiUrl: 'http://192.168.99.100:30090/api/v1/',
  monitorApiUrl: 'http://192.168.99.100:30080/service/console-api/',
  documentRootUrl: 'https://developer.lightbend.com/docs/enterprisesuiteconsole/beta/user-guide/index.html',
  grafanaUrl: `http://192.168.99.100:30080/service/grafana/dashboard/script/exporter-async.js`,
  minLogLevel: 'info',
  defaultHttpTimeoutMillis: 10000
};
