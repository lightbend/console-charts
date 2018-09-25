export class Navigation {
    static goClusterPage() {
        cy.visit('/');
        cy.url().should('be', '/clusters');
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

    static goMonitorPage(workloadId: string, monitorId: string) {
        cy.visit(`/workloads/${workloadId}/monitors/${monitorId}`);
    }


}
