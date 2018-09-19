Cypress.Commands.add('goWorkloadPage', (value: string) => {
    cy.log(`go to workload page ${value}`);
    cy.visit('/');
    cy.get('rc-workload-table').contains(value).click();
});

Cypress.Commands.add('setFormGroupBy', (value: string) => {
    cy.log(`set form groupby to ${value}`);

    // FIXME: long wait time in create monitor mode (maybe related to use local-prom to get labels)
    cy.get('#agg-label option', {timeout: 20000}).contains(value);
    cy.wait(1000); // make sure downdown updated and no flaky
    cy.get('#agg-label').select(value);
});


// see more example of adding custom commands to Cypress TS interface
// in https://github.com/cypress-io/add-cypress-custom-command-in-typescript
// add new command to the existing Cypress interface
// tslint:disable-next-line no-namespace
declare namespace Cypress {
    // tslint:disable-next-line interface-name
    interface Chainable {
        goWorkloadPage: (value: string) => void;
        setFormGroupBy: (value: string) => void;
    }
  }

// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add("login", (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add("drag", { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add("dismiss", { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This is will overwrite an existing command --
// Cypress.Commands.overwrite("visit", (originalFn, url, options) => { ... })
