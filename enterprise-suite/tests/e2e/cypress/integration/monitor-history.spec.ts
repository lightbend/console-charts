import { Action } from '../support/action';
import { Form } from '../support/form';
import { Util } from '../support/util';
import { Navigation } from '../support/navigation';
import { History } from '../support/history';

// TODO: roll back history test

describe('History Log Test', () => {
  beforeEach(() => {
    // Util.deleteMonitorsForWorkload('console-frontend');
  });

  it('log history check', () => {
    // create and save monitor
    const monitorName = Util.createRandomMonitorName();
    Navigation.goWorkloadPageByClick('console-frontend');
    Action.createMonitor();
    Form.setMetricName('kube_pod_failed');
    Form.setMonitorName(monitorName);
    Action.saveMonitor();

    // go to monitor page
    Util.validateUrlPath('/namespaces/lightbend/workloads/console-frontend');
    Util.validateMonitorCountGte(3);
    Navigation.clickMonitorByName(monitorName);
    Util.validateUrlPath(`/namespaces/lightbend/workloads/console-frontend/monitors/${monitorName}`);

    // check history log: should be 1 item
    cy.log('check log history when create a new monitor');
    History.validateCount(1);
    History.validateCreatedIsIndex(0);
    Action.editMonitor();
    Form.validateGroupByNone();

    // save monitor
    Form.setGroupBy('instance');
    Form.setAggregateUsing('avg');
    Action.saveMonitor();

    // go to monitor page and check history
    Util.validateUrlPath('/namespaces/lightbend/workloads/console-frontend');
    Util.validateMonitorCountGte(3);
    Navigation.clickMonitorByName(monitorName);
    Util.validateUrlPath(`/namespaces/lightbend/workloads/console-frontend/monitors/${monitorName}`);

    History.validateCreatedIsIndex(1);
    History.validateModifiedIsIndex(0);
    History.validateCount(2);

    // ISSUE: lightbend/console-frontend#501 - Monitor Change Log is missing details
    if (!Cypress.env('skipKnownError')) {
      // FIXME: the following test is disabled due to regression failed
      History.validateContainChange(0, 'aggregate using', 'avg');
      History.validateContainChange(0, 'group by', 'instance');
      // FIXME: bug in frontend, should only show two lines for first change
      // (1) title 'modified' (2) modified content
      History.validateChangeCountForIndex(0, 2);
    }

    // clean up
    Action.editMonitor();
    Action.removeMonitor();
    Util.validateUrlPath('/namespaces/lightbend/workloads/console-frontend');
    Util.validateMonitorCountGte(3);
    Util.validateNoMonitor(monitorName);
  });

});
