export class Navigation {
    static goClusterPage() {
        cy.visit('/cluster');
        cy.url().should('be', '/cluster');
    }

    static clickWorkload(workloadId: string) {
        cy.get(`rc-workload-table .workload-row[workloadname="${workloadId}"]`, {timeout: 10000}).click();
        cy.url().should('include', `/workloads/${workloadId}`);
    }

    static goWorkloadPageByClick(workloadId: string) {
        this.goClusterPage();
        this.clickWorkload(workloadId);
    }

    static clickMonitor(monId: string) {
        cy.get(`.monitor-list .monitor-name[title="${monId}"]`, {timeout: 12000}).click();
    }

    static clickMonitorByName(monId: string) {
        cy.get(`.monitor-list .monitor-name[id="${monId}"]`, {timeout: 12000}).click();
    }

    static goMonitorPage(namespace: string, workloadId: string, monitorId: string) {
        cy.visit(`/namespaces/${namespace}/workloads/${workloadId}/monitors/${monitorId}`);
    }


}
