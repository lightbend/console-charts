# helm-charts

Contains public Helm charts for various Lightbend projects. This project is hosted on [GitHub Pages](https://lightbend.github.io/helm-charts/index.yaml).

## Helm install enterprise suite:

```bash
kubectl create serviceaccount --namespace kube-system tiller

kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller

helm init --service-account tiller

helm repo add lightbend-helm-charts https://lightbend.github.io/helm-charts

helm repo update

helm install lightbend-helm-charts/enterprise-suite --name=es --namespace=lightbend --debug
```

## kubectl apply enterprise suite:

```
kubectl create namespace lightbend

kubectl --namespace=lightbend apply -f https://lightbend.github.io/helm-charts/es/all.yaml

```

## Publishing Charts

_wip_

Make any changes to the chart, commit it and open a PR. Once the change is merged to master,
you can release it directly:

    scripts/make-release.sh <version> <chart-name>
    git push --tags
    
For example:

    scripts/make-release.sh 0.0.15 enterprise-suite
    git push --tags

This will set the chart version, package it, and make a
commit. Finally, `git push --tags` will publish the release and git tag.

## Development

A `Makefile` is provided with useful targets for development.

* `make` to check the chart for errors, and update `docs` and `resources`.
* `make install-local` to try the release in docs.
* `make lint` if you just want to check for errors.

To specify the chart to build, set the `CHART` variable:

    make CHART=reactive-suite
    make CHART=enterprise-suite
    
By default `CHART=enterprise-suite`.

## Maintenance

Enterprise Suite Team <es-all@lightbend.com>

## License

Copyright (C) 2017 Lightbend Inc. (https://www.lightbend.com).

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this project except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
