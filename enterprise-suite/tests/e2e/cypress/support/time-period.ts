export class TimePeriod {
    static select(timePeriod: 'last hr'|'last 4 hrs'|'last day'|'last week') {
        cy.get('.time-period-select').find('option').should('have.length', 4);
        cy.get('.time-period-select').select(timePeriod);
    }
}
