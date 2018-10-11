import { Action } from '../support/action';
import { Form } from '../support/form';
import { Util } from '../support/util';
import { Navigation } from '../support/navigation';
import { History } from '../support/history';

// TODO: roll back history test

describe('History Log Test', () => {
  beforeEach(() => {
    // Util.deleteMonitorsForWorkload('es-console');
  });

  it.skip('log history check', () => {
    // create and save monitor
    const monitorName = Util.createRandomMonitorName();
    Navigation.goWorkloadPageByClick('es-console');
    Action.createMonitor();
    Form.setMetricName('kube_pod_failed');
    Form.setMonitorName(monitorName);
    Action.saveMonitor();

    // go to monitor page
    Util.validateUrlPath('/workloads/es-console');
    Util.validateMonitorCountGte(3);
    Navigation.clickMonitor(monitorName);
    Util.validateUrlPath(`/workloads/es-console/monitors/${monitorName}`);

    // check history log: should be 1 item
    cy.log('check log history when create a new monitor');
    if (!Cypress.env('skipKnownError')) {
      History.validateCount(1);
      History.validateCreatedIsIndex(0);
    }
    Action.editMonitor();
    Form.validateGroupByNone();

    // save monitor
    Form.setGroupBy('es_workload');
    Form.setAggregateUsing('avg');
    Action.saveMonitor();

    // go to monitor page and check history
    Util.validateUrlPath('/workloads/es-console');
    Util.validateMonitorCountGte(3);
    Navigation.clickMonitor(monitorName);
    Util.validateUrlPath(`/workloads/es-console/monitors/${monitorName}`);


    History.validateCreatedIsIndex(1);
    History.validateModifiedIsIndex(0);
    History.validateContainChange(0, 'aggregate using', 'avg');
    History.validateContainChange(0, 'group by', 'es_workload');
    // ISSUE: lightbend/console-home#260 - "filter by" field should not change if only change group by
    if (!Cypress.env('skipKnownError')) {
      History.validateCount(2);
      // FIXME: bug in frontend, should only show two changes
      History.validateChangeCountForIndex(0, 2);
    }

    // clean up
    Action.editMonitor();
    Action.removeMonitor();
    Util.validateUrlPath('/workloads/es-console');
    Util.validateMonitorCountGte(3);
    Util.validateNoMonitor(monitorName);
  });

});
