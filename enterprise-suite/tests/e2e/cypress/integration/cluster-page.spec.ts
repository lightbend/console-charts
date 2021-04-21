import { Navigation } from '../support/navigation';
import { ClusterDetails } from '../support/cluster-details';
import { ClusterPage } from '../support/cluster-page';

describe('Cluster Page Test', () => {
  beforeEach(() => {
    Navigation.goClusterPage();
  });

  it('cluster detail panel check', () => {
    ClusterDetails.infraContains('Nodes', '1');
    ClusterDetails.workloadHealthGte('Healthy', 10);
  });

  it('workload list check', () => {
    ClusterPage.validateWorkloadCountGte(10);
    ClusterPage.validateNodePodContainerCount('es-demo', 1, 3, 3);
  });

  it('switch namespace check', () => {
    ClusterPage.switchNamespace('lightbend');
    ClusterPage.validateWorkloadCountGte(5);
    ClusterPage.validateWorkloadCountLte(10);
  });

});
