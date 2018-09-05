[![Build Status](https://travis-ci.org/lightbend/helm-charts.svg?branch=master)](https://travis-ci.org/lightbend/helm-charts)

# helm-charts

Contains public Helm charts for various Lightbend projects. This project is hosted on [GitHub Pages](https://lightbend.github.io/helm-charts/index.yaml).

## Project layout

All projects must conform to the [Helm chart layout](https://docs.helm.sh/developing_charts/).

If there are files that you don't want included in the chart, add them to a `.helmignore` file in your project directory.

## Project Makefile

All projects must have a `Makefile` that implements two targets:

- `package`:  This target should create the chart tarball and push it up to the `helm-charts/docs` directory.
- `test`:  This target is run by Travis to test the release.

A default `common.mk` file is included that can be used for this purpose, although a project is free to implement these targets as they see fit.

## Helm-charts Makefile

The default target of the top-level Makefile builds the `index.yaml` file based on the tarballs.

If either of the `package` or `test` targets are invoked, they are recursively invoked on each of the projects.  To run over a particular subset of projects just define a value on the command line.  e.g. `make package CHARTS=sample-project`

## Helm install a project:

To install the Helm chart for a particular project...

```bash
kubectl create serviceaccount --namespace kube-system tiller

kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller

helm init --service-account tiller

helm repo add lightbend-helm-charts https://lightbend.github.io/helm-charts

helm repo update

helm install lightbend-helm-charts/<chart-name> --name=<release-name> --namespace=lightbend --debug
```

### Upgrade

```
helm repo update
helm upgrade lightbend-helm-charts/<chart-name> --name=<release-name> --namespace=lightbend --debug
```

## Cutting a Release / Publishing Charts

_This needs updating..._

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

## Maintenance

Enterprise Suite Team <es-all@lightbend.com>

## License

Copyright (C) 2018 Lightbend Inc. (https://www.lightbend.com).

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this project except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
