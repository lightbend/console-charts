import { Util } from '../support/util';
import { Environment } from '../support/environment';

describe('Grafana Test', () => {
  const grafanaUrl = Environment.getEnv().grafanaUrl +
    '?es_workload=es-demo&from=now-4h&service-type=akka,kubernetes&' +
    'metric=akka_actor_processing_time_ns&monitor=akka_processing_time&' +
    'promQL=max without (es_monitor_type) (akka_actor_processing_time_ns{quantile="0.5",es_workload="es-demo"})';

  it('open grafana url in monitor page', () => {
    cy.visit('/workloads/es-demo/monitors/akka_processing_time', {
      onBeforeLoad(win) {
        cy.stub(win, 'open').as('windowOpen'); // stub window.open event
      }
    });
    Util.validateControlIconContains('Grafana');
    Util.clickControlIcon('Grafana');
    cy.get('@windowOpen').should('be.calledWith', grafanaUrl);
  });

  it('no grafana graph error after open it from monitor page', () => {
    cy.visit(grafanaUrl);
    cy.contains('Monitored Metrics');
    cy.contains('Akka Metrics');
    cy.contains('Kubernetes Metrics');
    cy.get('.graph-panel__chart canvas', {timeout: 6000}); // draw canvas
    cy.get('.panel-info-corner--error').should('have.length', 0); // no error graph
  });

});
