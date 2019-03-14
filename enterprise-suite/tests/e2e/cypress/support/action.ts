export class Action {
    static createMonitor() {
        cy.contains('CREATE MONITOR').click();
    }

    static saveMonitor() {
        cy.contains('SAVE CHANGES').click();
    }

    static editMonitor() {
        cy.get('.form-container').contains('EDIT').should('not.have.attr', 'hidden');

        cy.get('.form-container').contains('EDIT').click();

        cy.get('.form-container').contains('EDIT').should('have.attr', 'hidden');
    }

    static removeMonitor() {
        cy.get('.form-container').contains('REMOVE MONITOR').click();
    }
}
