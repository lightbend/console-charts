import { Navigation } from '../support/navigation';
import { WorkloadDetails } from '../support/workload-details';
import { Util } from '../support/util';

describe('Workload Page Test', () => {
  beforeEach(() => {
    Navigation.goWorkloadPageByClick('es-demo');
  });

  it('validate detail panels', () => {
    Util.validateMonitorCountGte(3);
    WorkloadDetails.validateNodePodContainerCount(1, 3, 3);
    WorkloadDetails.validateServiceType(['akka', 'kubernetes']);
    WorkloadDetails.validateLabelsContains('namespace', 'default');
    WorkloadDetails.validateLabelsContains('node_name', 'minikube');
  });

  it('validate control icons', () => {
    Util.validateControlIconContains('Grafana');
    Util.validateControlIconContains('Kubernetes');
  });

});
