# helm-charts/enterprise-suite

Helm chart setup for the Enterprise Suite project.

(Note that this README is not included in the Chart package.)

## Helm install enterprise suite:

See ../README.md for general instructions for installing chart as these are applicable to the installation of Enterprise Suite as well.
In the normal case use

```bash
helm install lightbend-helm-charts/enterprise-suite --name=es --namespace=lightbend --debug
```

To run Enterprise Suite in a minikube environment, instead use

```bash
helm install lightbend-helm-charts/enterprise-suite --name=es --namespace=lightbend --debug --set minikube=true
```

## "latest" internal dev release:

By default all the helm charts use versioned images, so you are using fixed dependencies.
Enterprise Suite also has a special "latest" chart which uses "latest" tags for images. This is
useful for development.

```bash
helm install lightbend-helm-charts/enterprise-suite-latest --name=es --namespace=lightbend --debug
```

To work with minikube, instead install with:

```bash
helm install lightbend-helm-charts/enterprise-suite-latest --name=es --namespace=lightbend --debug --set minikube=true
```

When PRs are merged to master, "latest" is updated automatically.  If you have installed the  `lightbend-helm-charts/enterprise-suite-latest` helm chart or  `es/all-latest.yaml` you should be able to simply delete the pod you want upgraded and minikube will go fetch the latest from bintray and install it for you.  You will only have to do a helm upgrade on this track if you want to pick up configuration changes in the helm chart itself.

### Upgrade

```
helm repo update
helm upgrade lightbend-helm-charts/enterprise-suite --name=es --namespace=lightbend --debug

# or upgrade chart with "latest" container images
helm upgrade lightbend-helm-charts/enterprise-suite-latest --name=es --namespace=lightbend --debug
```

If using minikube, include the following option in the `upgrade` command:
```
set minikube=true
```

## kubectl apply enterprise suite:

If you want to use `kubectl` to apply the enterprise suite
resource definitions then follow the [instructions in the official documentation](https://developer.lightbend.com/docs/console/current/installation/es.html#use-the-install-subcommand-to-generate-the-resource-yaml).

## Cutting a Release / Publishing Charts

See the [Google doc for releasing Enterprise Suite
Console](https://docs.google.com/document/d/14L3Zdwc-MkCDR1-7fWQYQT3k53vLc4cehAKEuOnwhxs)
for the definitive process.

## Development

A `Makefile` is provided with useful targets for development.

* `make` to check the chart for errors, and update `docs` and `resources`.
* `make install-local` to try the release in docs.
* `make lint` if you just want to check for errors.
* `make test` to run end-to-end tests against a local minikube. This requires that minikube and helm clients are installed.

## Maintenance

Enterprise Suite Team <es-all@lightbend.com>

## License

Copyright (C) 2017 Lightbend Inc. (https://www.lightbend.com).

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this project except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
