export class Health {
    /** check the last color in middle health bar */
    // This code is already skipped from caller due to issue #353 and #354
    static validateMiddleMetricList(index: number, health: 'warning'|'critical'|'ok') {
        cy.log('validateMiddleMetricList()', index, health);

        // ISSUE: lightbend/console-home#353 - health bar recalculate twice, so put long delay to bypass problem
        // cy.get('.monitor-list .health-bar')
        //     .eq(index)
        //     .find('rect');
        // cy.wait(10000); // wait health bar be stable, need to remove in the future
        cy.get('.monitor-list .health-bar')
            .eq(index)
            .find('rect')
            .not('.crosshair')
            .last({timeout: 10000})
            // ISSUE/FLAKY: lightbend/console-home#354 - unknown health in middle health bar due to missing health data
            .should('have.class', `health-${health}-bar`);
    }

    static validateSelectedGraph(health: 'warning'|'critical'|'ok') {
        cy.log('validateSelectedGraph()', health);
        cy.get('.selected-container .health-bar')
            .find('rect')
            .not('.crosshair')
            .last({timeout: 10000})
            .should('have.class', `health-${health}-bar`);
    }

    static validateBottomTimeline(health: 'warning'|'critical'|'ok') {
        cy.log('validateBottomTimeline()', health);
        cy.wait(5000); // FIXME hard wait
        cy.get('.timeline-health')
            .find('rect')
            .not('.crosshair')
            .last({timeout: 10000})
            .should('have.class', `health-${health}-bar`);
    }
}
