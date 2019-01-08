export class ClusterDetails {
    static infraContains(key: string, value: string) {
        cy.get('rc-panel[title="Infrastructure"]', { timeout: 10000 })
            .contains('.label-key', key, {timeout: 10000})
            .parent()
            .contains('.label-value', value, {timeout: 10000});
    }

    static workloadHealthGte(severity: 'Healthy'|'Warning'|'Critical'|'Unknown', count: number) {
        cy.get('rc-panel[title="Workload Health"]').contains(severity)
            .parent().children('.label-value').find('.right-count')
            .then(($dom) => {
                const realCount = Number($dom.text());
                expect(realCount).to.gte(count);
            }
        );
    }


}
