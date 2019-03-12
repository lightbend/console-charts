import { Action } from '../support/action';
import { Form, ThresholdMonitor } from '../support/form';
import { Util } from '../support/util';
import { Navigation } from '../support/navigation';
import { Health } from '../support/health';

// NB, disabling all tests here for now, per the resolution of https://github.com/lightbend/console-home/issues/328

describe('Local Prometheus Health Test', () => {
  beforeEach(() => {
    Navigation.goMonitorPage('lightbend', 'es-demo', 'akka_inbox_growth');
    Form.validateMonitorName('akka_inbox_growth');
    Action.editMonitor();
    Form.validateMonitorName('akka_inbox_growth');
    Form.validateMonitorType('growth rate');
    Form.validateMetricName('akka_actor_mailbox_size');
  });

  it.skip('enable critical only', () => {
    const thresholdMonitor: ThresholdMonitor = {
      groupBy: 'actor',
      timeWindow: '1 minute',
      triggerOccurrence: 'once',
      critical: {enabled: true, comparator: '<', value: 3},
      warning: {enabled: false, comparator: '<', value: 5},
      aggregator: 'max'
    };

    Form.setThresholdMonitor(thresholdMonitor);
    Util.waitRecalculateHealth();
    Health.validateMiddleMetricList(0, 'critical');
    Health.validateMiddleMetricList(1, 'critical');
    Health.validateSelectedGraph('critical');
    Health.validateContextTimeline('critical');
  });

  it.skip('enable warning only', () => {
    const thresholdMonitor: ThresholdMonitor = {
      groupBy: 'actor',
      timeWindow: '1 minute',
      triggerOccurrence: 'once',
      critical: {enabled: false, comparator: '<', value: 3},
      warning: {enabled: true, comparator: '<', value: 5},
      aggregator: 'max'
    };

    Form.setThresholdMonitor(thresholdMonitor);
    Util.waitRecalculateHealth();
    Health.validateMiddleMetricList(0, 'warning');
    Health.validateMiddleMetricList(1, 'warning');
    Health.validateSelectedGraph('warning');
    Health.validateContextTimeline('warning');
  });

  it.skip('disable both critical and warning', () => {
    const thresholdMonitor: ThresholdMonitor = {
      groupBy: 'actor',
      timeWindow: '1 minute',
      triggerOccurrence: 'once',
      critical: {enabled: false, comparator: '<', value: 3},
      warning: {enabled: false, comparator: '<', value: 5},
      aggregator: 'max'
    };

    Form.setThresholdMonitor(thresholdMonitor);
    Util.waitRecalculateHealth();
    // FIXME: warning should not overwrite critical
    Health.validateMiddleMetricList(0, 'ok');
    Health.validateMiddleMetricList(1, 'ok');
    Health.validateSelectedGraph('ok');
    Health.validateContextTimeline('ok');
  });

  // Disable slow test
  it.skip('warning should not overwrite critical', () => {
    const thresholdMonitor: ThresholdMonitor = {
      groupBy: 'actor',
      timeWindow: '1 minute',
      triggerOccurrence: 'once',
      critical: {enabled: true, comparator: '<', value: 3},
      warning: {enabled: false, comparator: '>', value: 1.2},
      aggregator: 'max'
    };

    Form.setThresholdMonitor(thresholdMonitor);

    // move the following to health check
    Util.validateMidHealthBarCount(2);
    Health.validateMiddleMetricList(0, 'critical');
    Health.validateMiddleMetricList(1, 'critical');
    Health.validateSelectedGraph('critical');
    Health.validateContextTimeline('critical');

    Form.enableWarning(true);
    Util.waitRecalculateHealth();

    Health.validateMiddleMetricList(0, 'critical');
    Health.validateMiddleMetricList(1, 'critical');
    Health.validateSelectedGraph('critical');
    Health.validateContextTimeline('critical');
  });

  it.skip('enable both critical and warning 1', () => {
    const thresholdMonitor: ThresholdMonitor = {
      groupBy: 'actor',
      timeWindow: '1 minute',
      triggerOccurrence: 'once',
      critical: {enabled: true, comparator: '<', value: 3},
      warning: {enabled: true, comparator: '>', value: 5},
      aggregator: 'max'
    };

    Form.setThresholdMonitor(thresholdMonitor);
    Util.waitRecalculateHealth();
    Health.validateMiddleMetricList(0, 'critical');
    Health.validateMiddleMetricList(1, 'critical');
    Health.validateSelectedGraph('critical');
    Health.validateContextTimeline('critical');
  });

  it.skip('enable both critical and warning 2', () => {
    const thresholdMonitor: ThresholdMonitor = {
      groupBy: 'actor',
      timeWindow: '1 minute',
      triggerOccurrence: 'once',
      critical: {enabled: true, comparator: '<', value: 3},
      warning: {enabled: true, comparator: '<', value: 5},
      aggregator: 'max'
    };

    Form.setThresholdMonitor(thresholdMonitor);
    Util.waitRecalculateHealth();
    Health.validateMiddleMetricList(0, 'critical');
    Health.validateMiddleMetricList(1, 'critical');
    Health.validateSelectedGraph('critical');
    Health.validateContextTimeline('critical');
  });
});
