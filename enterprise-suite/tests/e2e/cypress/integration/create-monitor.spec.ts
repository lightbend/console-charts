import { Action } from '../support/action';
import { Form, ThresholdMonitor } from '../support/form';
import { Util } from '../support/util';
import { Navigation } from '../support/navigation';

// TODO: need to validate convert-to-rate in the future

describe('Create Monitor Test', () => {

  beforeEach(() => {
    // Util.deleteMonitorsForWorkload('es-demo');
  });


  // NOTE: group by actor is flaky due to the following two issues
  it('validate created threshold monitor w/ group by actor', () => {
    // create and save monitor
    const monitorName = Util.createRandomMonitorName();
    Navigation.goWorkloadPageByClick('es-demo');
    Action.createMonitor();

    const thresholdMonitor: ThresholdMonitor = {
      monitorName: monitorName,
      metric: 'akka_actor_mailbox_size',
      groupBy: 'actor',
      timeWindow: '1 minute',
      triggerOccurrence: 'once',
      critical: {enabled: true, comparator: '<', value: 3},
      warning: {enabled: false, comparator: '>', value: 1.2},
      aggregator: 'max'
    };

    Form.setThresholdMonitor(thresholdMonitor);
    Form.addFilterBy('app', 'es-demo');
    Action.saveMonitor();

    // go to monitor page
    Util.validateUrlPath('/namespaces/lightbend/workloads/es-demo');
    Util.validateMonitorCountGte(3);

    // go to created monitor
    Navigation.clickMonitor(monitorName);
    Util.validateUrlPath(`/namespaces/lightbend/workloads/es-demo/monitors/${monitorName}`);

    // validate form
    Action.editMonitor();
    Form.validateThresholdMonitor(thresholdMonitor);
    Form.validateFilterByCount(1);
    Form.validateFilterByContains('app', 'es-demo');

    // delete created monitor
    Action.removeMonitor();
    Util.validateUrlPath(`/namespaces/lightbend/workloads/es-demo`);
    Util.validateMonitorCountGte(3);
    Util.validateNoMonitor(monitorName);
  });

  // ISSUE/FLAKY: lightbend/console-home#323 - sometimes drop down data is incorrect
  it.skip('validate created threshold monitor(basic)', () => {
    // create and save monitor
    const monitorName = Util.createRandomMonitorName();
    Navigation.goWorkloadPageByClick('es-demo');
    Action.createMonitor();

    const thresholdMonitor: ThresholdMonitor = {
      monitorName: monitorName,
      metric: 'kube_pod_failed',
      groupBy: 'pod',
      timeWindow: '1 minute',
      triggerOccurrence: 'once',
      critical: {enabled: true, comparator: '<', value: 3},
      warning: {enabled: false, comparator: '>', value: 1.2},
      aggregator: 'max'
    };

    Form.setThresholdMonitor(thresholdMonitor);
    Form.addFilterBy('job', 'kube-state-metrics');
    Action.saveMonitor();

    // go to monitor page
    Util.validateUrlPath('/namespaces/lightbend/workloads/es-demo');
    Util.validateMonitorCountGte(3);

    // go to created monitor
    Navigation.clickMonitor(monitorName);
    Util.validateUrlPath(`/namespaces/lightbend/workloads/es-demo/monitors/${monitorName}`);

    // validate form
    Action.editMonitor();
    Form.validateThresholdMonitor(thresholdMonitor);
    Form.validateFilterByCount(1);
    Form.validateFilterByContains('job', 'kube-state-metrics');

    // delete created monitor
    Action.removeMonitor();
    Util.validateUrlPath(`/namespaces/lightbend/workloads/es-demo`);
    Util.validateMonitorCountGte(3);
    Util.validateNoMonitor(monitorName);
  });

});
