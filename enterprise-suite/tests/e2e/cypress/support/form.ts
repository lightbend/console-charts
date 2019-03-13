type Comparator = '='|'!='|'>'|'<'|'>='|'<=';
type Aggregator = 'avg'|'min'|'max'|'sum';
type TimeWindow = '1 minute'|'5 minutes'|'10 minutes'|'15 minutes'|'30 minutes'|'1 hour'|'2 hours'| '4 hours';
type Occurrence = 'once'|'25%'|'50%'|'75%'|'95%'|'100%';
type MonitorType = 'threshold'|'simple moving average'|'growth rate';

interface Severity {
    enabled: boolean;
    comparator: Comparator;
    value: number;
}

export interface ThresholdMonitor {
    monitorName?: string;
    metric?: string;
    groupBy: string;
    timeWindow: TimeWindow;
    triggerOccurrence: Occurrence;
    critical: Severity;
    warning: Severity;
    aggregator: Aggregator;
    // not put filter
    // not put convert-to-rate
}

export class Form {
    static clickMetricSelector() {
        cy.get('rc-capsule.metric').click();
    }

    static setGroupBy(value: string) {
        cy.log('set group by', value);
        cy.get(`#agg-label option[value="${value}"]`, {timeout: 40000});
        cy.wait(2000); // make sure downdown updated and no flaky
        cy.get('#agg-label', {timeout: 40000}).select(value);
    }

    static validateGroupBy(value: string) {
        cy.get('.form-container #agg-label').should('have.value', value);
    }

    static validateGroupByNone() {
        cy.get('.form-container #agg-label').should('have.value', '__none__');
        cy.get('.form-container .label').should('not.contain', 'Aggregate Using');
    }

    static setMetricName(value: string) {
        // assume dropdown is expanding
        cy.get('rc-capsule.metric .capsule-view').should('have.attr', 'hidden');
        cy.get('rc-capsule.metric .capsule-edit').should('not.have.attr', 'hidden');
        cy.get('rc-capsule.metric .capsule-edit .tag-name-input').clear().wait(1000).should('have.value', '').wait(1000).type(value);
        cy.wait(2000);
        cy.get(`rc-capsule.metric .capsule-edit .capsule-wrapper .tag-name-list a[title="${value}"]`, {timeout: 60000}).click();
        cy.wait(2000);
        this.validateMetricName(value);
    }

    static validateMetricName(value: string) {
        cy.get('rc-capsule.metric .capsule-view').should('not.have.attr', 'hidden');
        cy.get('rc-capsule.metric .capsule-edit').should('have.attr', 'hidden');
        cy.get('rc-capsule.metric .capsule-view label.button-key', {timeout: 20000}).should('have.text', value)
    }


    static setMonitorName(value: string) {
        cy.get('#mon-name').clear().wait(500).should('have.value', '').wait(500).type(value);
    }

    static validateMonitorName(value: string) {
        cy.get('#mon-name', {timeout: 5000}).should('have.value', value);
    }

    static setMonitorType(value: MonitorType) {
        cy.get('#monitor-type').select(value);
        cy.get('#monitor-type').should('have.value', value);
    }

    static validateMonitorType(value: MonitorType) {
      const mapping = {
        'growth rate': 'growthrate',
        'threshold': 'threshold',
        'simple moving average': 'sma'
      }
        cy.get('#monitor-type').should('have.value', mapping[value]);
    }

    static setTriggerOccurrence(value: Occurrence) {
        cy.get('#trigger-at-least').select(value);
    }

    static validateTriggerOccurrence(value: Occurrence) {
        const mapping = {
          'once': '5e-324',
          '25%': '0.25',
          '50%': '0.50',
          '75%': '0.75',
          '95%': '0.95',
          '100%': '1'
        }
        cy.get('#trigger-at-least').should('have.value', mapping[value]);
    }

    static setTimeWindow(value: TimeWindow) {
        cy.get('#trigger-within', {timeout: 5000}).select(value);
    }

    static validateTimeWindow(value: TimeWindow) {
        cy.get('#trigger-within').find(':selected').should('have.text', value);
    }

    static enableCritical(enable: boolean) {
        if (enable) { // enable
            cy.get('rc-ui-switch.critical-enable .fas').then(($switch) => {
                if ($switch.hasClass('fa-toggle-off')) {
                    $switch.click();
                }
            });

            cy.get('rc-ui-switch.critical-enable i').should('have.class', 'fa-toggle-on');

        } else { // disable
            cy.get('rc-ui-switch.critical-enable .fas').then(($switch) => {
                if ($switch.hasClass('fa-toggle-on')) {
                    $switch.click();
                }
            });

            cy.get('rc-ui-switch.critical-enable i').should('have.class', 'fa-toggle-off');
        }
    }

    static enableWarning(enable: boolean) {
        if (enable) { // enable
            cy.get('rc-ui-switch.warning-enable .fas').then(($switch) => {
                if ($switch.hasClass('fa-toggle-off')) {
                    $switch.click();
                }
            });
            cy.get('rc-ui-switch.warning-enable i').should('have.class', 'fa-toggle-on');
        } else { // disable
            cy.get('rc-ui-switch.warning-enable .fas').then(($switch) => {
                if ($switch.hasClass('fa-toggle-on')) {
                    $switch.click();
                }
            });
            cy.get('rc-ui-switch.warning-enable i').should('have.class', 'fa-toggle-off');
        }
    }

    static setCritical(comparator: Comparator, value: number) {
        cy.get('#critical-comparator').select(comparator);
        cy.get('#critical-threshold').clear().should('have.text', '').type(value.toString());
    }

    static setWarning(comparator: Comparator, value: number) {
        cy.get('#warning-comparator').select(comparator);
        cy.get('#warning-threshold').clear().should('have.text', '').type(value.toString());
    }

    static validateCritical(enabled: boolean, comparator: Comparator, value: number) {
        cy.get('.critical-enable > span').should('have.text', enabled ? 'Enabled' : 'Disabled');
        cy.get('#critical-comparator').find(':selected').should('have.text', comparator);
        cy.get('#critical-threshold').should('have.value', value.toString());
    }

    static validateWarning(enabled: boolean, comparator: Comparator, value: number) {
        cy.get('.warning-enable > span').should('have.text', enabled ? 'Enabled' : 'Disabled');
        cy.get('#warning-comparator').find(':selected').should('have.text', comparator);
        cy.get('#warning-threshold').should('have.value', value.toString());
    }

    static setAggregateUsing(value: Aggregator) {
        cy.get('#aggregation').select(value);
    }

    static validateAggregateUsing(value: Aggregator) {
        cy.get('#aggregation').should('have.value', value);
    }

    static addFilterBy(key: string, value: string) {
        cy.get('.form-container button.add').click();
        cy.get(`.form-container .tag-name-list a[title="${key}"]`).click();
        cy.get(`.form-container .tag-value-list a[title="${value}"]`).click();
    }

    static validateFilterByCount(value: number) {
        cy.get('.capsule-group rc-capsule').should('have.length', value);
    }

    static validateFilterByContains(key: string, value: string) {
        cy.get(`.capsule-group rc-capsule .capsule-view > .button-key`).contains(key);
        cy.get(`.capsule-group rc-capsule .capsule-view > .button-value`).contains(value);
    }

    static setThresholdMonitor(m: ThresholdMonitor) {
        if (m.metric) {
          this.setMetricName(m.metric);
        }

        if (m.monitorName) {
          this.setMonitorName(m.monitorName);
        }

        this.setMonitorType('threshold');
        cy.wait(2000); // work around ui change slowly
        this.setMonitorType('threshold'); // FIXME: sometimes it will roll back to growthrate monitor
        cy.wait(1000); // work around ui change slowly
        this.setTimeWindow(m.timeWindow);
        this.enableCritical(m.critical.enabled);
        this.setCritical(m.critical.comparator, m.critical.value);
        this.enableWarning(m.warning.enabled);
        this.setWarning(m.warning.comparator, m.warning.value);
        this.setTriggerOccurrence(m.triggerOccurrence); // new added
        this.setGroupBy(m.groupBy);
        if (m.groupBy !== '<none>') {
            this.setAggregateUsing(m.aggregator);
        }
    }

    static validateThresholdMonitor(m: ThresholdMonitor) {
        if (m.metric) {
          this.validateMetricName(m.metric);
        }

        if (m.monitorName) {
          this.validateMonitorName(m.monitorName);
        }

        this.validateMonitorType('threshold');
        this.validateTimeWindow(m.timeWindow);
        this.validateCritical(m.critical.enabled, m.critical.comparator, m.critical.value);
        this.validateWarning(m.warning.enabled, m.warning.comparator, m.warning.value);
        this.validateTriggerOccurrence(m.triggerOccurrence);
        if (m.groupBy !== '<none>') {
            this.validateGroupBy(m.groupBy);
            this.validateAggregateUsing(m.aggregator);
        } else {
            this.validateGroupByNone();
        }
    }

}
