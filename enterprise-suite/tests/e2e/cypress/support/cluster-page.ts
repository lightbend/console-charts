export class ClusterPage {
    static validateWorkloadCountGte(count: number) {
        cy.get('rc-workload-table .workload-row', {timeout: 10000}).should('have.length.be.gte', count);
    }

    static validateWorkloadCountLte(count: number) {
        cy.get('rc-workload-table .workload-row', {timeout: 10000}).should('have.length.be.lte', count);
    }

    static validateNodePodContainerCount(workload: string, nodeCount: number, podCount: number, containerCount: number) {
        cy.get(`[workloadname="${workload}"] > :nth-child(4)`)
            .should('have.text', `${nodeCount} : ${podCount} : ${containerCount}`);
    }

    static switchNamespace(namespace: string) {
        cy.get('select.namespace').select(namespace);
    }


}
