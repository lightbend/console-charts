export class WorkloadDetails {
  static validateNodePodContainerCount(nodeCount: number, podCount: number, containerCount: number) {
    cy.get('rc-workload-details [title="Infrastructure History"] > .panel-body > :nth-child(1) > svg > .last-value')
      .should('have.text', nodeCount.toString());
    cy.get('rc-workload-details [title="Infrastructure History"] > .panel-body > :nth-child(2) > svg > .last-value')
      .should('have.text', podCount.toString());
    cy.get('rc-workload-details [title="Infrastructure History"] > .panel-body > :nth-child(3)  > svg > .last-value')
      .should('have.text', containerCount.toString());
  }

  static validateServiceType(serviceTypes: string[]) {
    const count = serviceTypes.length;
    cy.get('rc-workload-details [title="Service Types"] .inset').should('have.length', count);
    serviceTypes.forEach(type => {
      cy.get('rc-workload-details [title="Service Types"] .inset').contains(type);
    });
  }

  static validateLabelsContains(key: string, value: string) {
    cy.get('rc-workload-details [title="Labels"]')
      .contains(key)
      .parent().children('.label-value')
      .should('have.text', value);
  }

}
