export class History {
    static validateCount(count: number) {
        cy.log('validate log history count');
        cy.get('side rc-monitor-change-log .circle').should('have.length', count);
    }

    static validateCreatedIsIndex(index: number) {
        cy.get(`side rc-monitor-change-log .log-type[data-index="${index}"] .value`).contains('Created');
        cy.get(`side rc-monitor-change-log .change-field[data-index="${index}"]`).should('have.length', 0);
    }

    static validateModifiedIsIndex(index: number) {
        cy.get(`side rc-monitor-change-log .log-type[data-index="${index}"] .value`).contains('Modified');
    }

    static validateChangeCountForIndex(index: number, count: number) {
        cy.get(`side rc-monitor-change-log .change-field[data-index="${index}"]`).should('have.length', count);
    }

    static validateContainChange(index: number, field: string, value: string) {
        cy.get(`side rc-monitor-change-log .change-field[data-index="${index}"] .field`).contains(field)
          .parent().children('.value').contains(value);
    }
}
