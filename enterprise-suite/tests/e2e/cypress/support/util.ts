import { Environment } from './environment';

export class Util {
  static deleteMonitorsForWorkload(workloadId: string) {
    const esMonitorApiServer = Environment.getEnv().monitorApiUrl;

    // delete monitors for workloadId
    cy.request({
      method: 'DELETE',
      url: `${esMonitorApiServer}/monitors/${workloadId}`,
      headers: {
          'Author-Name': 'Regression',
          'Author-Email': 'regression@gmail.com',
          'Message': 'testing!'
      },
      failOnStatusCode: false // not consider 500/404 as error
    }).then((response) => {
      expect(response.status).to.gte(200);
    });
  }

  static createRandomMonitorName() {
    const random = Math.floor(Math.random() * 100000);
    return `regression_test_${random}`;
  }

  static validateMidHealthBarCount(count: number) {
    cy.get('.list-header', {timeout: 20000}).should('have.text', `Groupings (${count})`);
  }

  static validateUrlPath(urlPath: string) {
    cy.location('pathname', {timeout: 20000}).should('eq', urlPath);
  }

  static validateMonitorCountGte(count: number) {
    cy.get('.monitor-list .monitor-name', {timeout: 10000}).should('have.length.be.gte', count);
  }

  static validateControlIconContains(value: string) {
      cy.get(`rc-cluster-controls img[title="${value}"]`, {timeout: 60000});
  }

  static clickControlIcon(value: string) {
      cy.get(`rc-cluster-controls img[title="${value}"]`).click();
  }

  static validateNoMonitor(monitorId: string) {
    cy.get('.monitor-list', {timeout: 10000}).should('not.contain', monitorId);
  }

  static waitRecalculateHealth() {
    cy.get('.monitor-list .center').contains('Calculating Health');
    cy.get('.monitor-list .health-bar', {timeout: 20000});
    cy.wait(5000); // FIXME: hard waiting to workaround update health line by line
  }

}
