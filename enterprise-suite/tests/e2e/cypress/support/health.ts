export class Health {
    /** check the last color in middle health bar */
    // This code is already skipped from caller due to issue #353 and #354
    static validateMiddleMetricList(index: number, health: 'warning'|'critical'|'ok') {
        cy.log('validateMiddleMetricList()', index, health);

        // FIXME: polling health bar is updated -> somehow code always have "health-unknown-bar", so we cannot poll This
        // Instead, if it has more then "2" rect (1) rect.health-unknown-bar (2) rect.corsshair  then it has updated
        cy.get('.monitor-list .health-bar').eq(index).find('rect', {timeout: 10000}).should('have.length.be.gt', 2);

        cy.wait(2000);  // FIXME: sometimes health bar updated is not stable. wait make it more stable
        cy.get('.monitor-list .health-bar')
            .eq(index)
            .find('rect')
            .not('.crosshair')
            .last({timeout: 2000})
            // ISSUE/FLAKY: lightbend/console-home#354 - unknown health in middle health bar due to missing health data
            .should('have.class', `health-${health}-bar`);
    }

    static validateSelectedGraph(health: 'warning'|'critical'|'ok') {
        cy.log('validateSelectedGraph()', health);
        cy.get('.selected-container .health-bar')
            .find('rect')
            .not('.crosshair')
            .last({timeout: 2000})
            .should('have.class', `health-${health}-bar`);
    }

    // second  timeline
    static validateContextTimeline(health: 'warning'|'critical'|'ok') {
        cy.log('validateContextTimeline()', health);

        // wait for health bar updated
        cy.get('.context-div .timeline-health').find('rect', {timeout: 10000}).should('have.length.be.gt', 2);

        cy.wait(2000); // FIXME: sometimes health bar updated is not stable. wait make it more stable

        cy.get('.context-div .timeline-health')
            .find('rect', {timeout: 2000})
            .not('.crosshair')
            // .last({timeout: 10000})  // NOTE: not the last element anymore
            .should('have.class', `health-${health}-bar`);
    }
}
