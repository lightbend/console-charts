import { base, Env } from './environment.base';

export const environment: Env = {
  ...base,
  production: true,
  browserProm: {
    enabled: true,
    debug: false
  },
  isConvertToRateEnabled: false,
  prometheusApiUrl: `service/prometheus/api/v1/`,
  monitorApiUrl: `service/console-api/`,
  grafanaUrl: `service/grafana/dashboard/script/exporter-async.js`,

  // TODO need nginx proxy_pass set up for this as well:
  kubernetesUrl: 'http://192.168.99.100:30000/',

  documentRootUrl: 'https://developer.lightbend.com/docs/enterprisesuiteconsole/beta/user-guide/index.html',
  minLogLevel: 'info',
  defaultHttpTimeoutMillis: 10000
};
