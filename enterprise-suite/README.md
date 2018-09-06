# helm-charts/enterprise-suite

Helm chart setup for the Enterprise Suite project.

(Note that this README is not included in the Chart package.)

## Helm install enterprise suite:

See ../README.md for instruction on how to install Enterprise Suite.

## "latest" internal dev release:

By default all the helm charts use versioned images, so you are using fixed dependencies.
There is also a special "latest" chart which uses "latest" tags for images. This is
useful for development.

```bash
helm install lightbend-helm-charts/enterprise-suite-latest --name=es --namespace=lightbend --debug
```

When PRs are merged to master, "latest" is updated automatically.  If you have installed the  `lightbend-helm-charts/enterprise-suite-latest` helm chart or  `es/all-latest.yaml` you should be able to simply delete the pod you want upgraded and minikube will go fetch the latest from bintray and install it for you.  You will only have to do a helm upgrade on this track if you want to pick up configuration changes in the helm chart itself.

### Upgrade

```
helm repo update
helm upgrade lightbend-helm-charts/enterprise-suite --name=es --namespace=lightbend --debug

# or upgrade chart with "latest" container images
helm upgrade lightbend-helm-charts/enterprise-suite-latest --name=es --namespace=lightbend --debug
```

## kubectl apply enterprise suite:

```
kubectl create namespace lightbend

kubectl --namespace=lightbend apply -f https://lightbend.github.io/helm-charts/es/all.yaml

# or use "latest" container images
kubectl --namespace=lightbend apply -f https://lightbend.github.io/helm-charts/es/all-latest.yaml
```

## Cutting a Release / Publishing Charts

_Is this still correct?_

### Release ES images

Release the images in Jenkins:
1. Go to <https://ci.lightbend.com/view/EntSuite/job/es-release-all/build>.
2. Click 'Build'. This will increment the version for each image and release it.
3. Open the [console](https://ci.lightbend.com/view/EntSuite/job/es-release-all/lastBuild/console), scroll to the bottom and you can see the versions of each image.
4. Update [enterprise-suite/values.yaml](enterprise-suite/values.yaml) with the new image versions and commit the changes.

See [es-build](https://github.com/lightbend/es-build) for more details.

### Release Charts

Install [yq](https://github.com/mikefarah/yq) if you don't have it yet:

    go get github.com/mikefarah/yq
    # or
    brew install yq                  

Then run the release script on a clean master checkout:

    scripts/make-release.sh enterprise-suite
    git push --follow-tags
    
This will increment the chart version, package it, and make a
commit. Finally, `git push --tags` will publish the release and git tag.

Optionally you can specify the version to use:

    scripts/make-release.sh enterprise-suite 1.0.0

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
