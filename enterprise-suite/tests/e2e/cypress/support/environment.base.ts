
export const unclassified = 'unclassified';
export const akka = 'akka';
export const akkaHttp = 'akka-http';
export const akkaFamily = 'akka family';
export const jvm = 'jvm';
export const kafka = 'kafka';
export const mongodb = 'mongodb';
export const nginx = 'nginx';
export const zookeeper = 'zookeeper';
export const kubernetes = 'kubernetes';
export interface Env {
  production: boolean;
  minLogLevel: 'trace' | 'info' | 'debug' | 'error';
  defaultHttpTimeoutMillis: number;
  isConvertToRateEnabled: boolean;
  isMutingEnabled: boolean;
  isMonitorEditEnabled: boolean;
  numDataSamples: number;
  defaultNS: string;
  browserProm: {
    enabled: boolean;
    debug: boolean
  };
  prometheusApiUrl: string;
  grafanaUrl: string;
  documentRootUrl: string;
  monitorApiUrl: string;
  kubernetesUrl: string;
  isRuleCreationEnabled: boolean; // backend does not support saved monitor rule yet.
  showClusterMap: boolean;
  graphTransitionDurationMS: number;
  // Dummy data
  isUsingDummyNewMonitorApiData: boolean;

  // Number of seconds after which to refresh daemonized data:
  dataRefreshInterval: number;
  logger: {
    '*': boolean;
    namespaces: { [key: string]: boolean };
    types: { [key: string]: boolean };
  };
  data: {
    maxResolution: {
      viewMode: number;
      editMode: number;
    }
  };
}


// TODO if a requirement is multiple service keys go to the same icon (of the same name), then this won't work
export const base = {
  production: false,
  isConvertToRateEnabled: true,
  isMutingEnabled: false,
  isMonitorEditEnabled: true,
  numDataSamples: 360,
  defaultNS: 'all',
  browserProm: {
    enabled: true,
    debug: true
  },
  grafanaUrl: 'http://192.168.99.100:30030/dashboard/script/exporter-async.js',
  monitorApiUrl: 'http://192.168.99.100:30080/service/console-api',
  kubernetesUrl: 'http://192.168.99.100:30000/',
  isRuleCreationEnabled: false, // backend does not support saved monitor rule yet.
  showClusterMap: true,
  graphTransitionDurationMS: 500,

  // Dummy data
  isUsingDummyNewMonitorApiData: false,

  // Number of seconds after which to refresh daemonized data:
  dataRefreshInterval: 10,
  logger: {
    '*': true,
    namespaces: {},
    types: {}
  },
  data: {
    maxResolution: {
      viewMode: 10,
      editMode: 10
    }
  }
};
