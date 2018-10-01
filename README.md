[![Build Status](https://travis-ci.org/lightbend/helm-charts.svg?branch=master)](https://travis-ci.org/lightbend/helm-charts)

# helm-charts

Contains public Helm charts for various Lightbend projects. The helm repo is available at `https://repo.lightbend.com/helm-charts`.

## Project layout

All projects must conform to the [Helm chart layout](https://docs.helm.sh/developing_charts/).

If there are files that you don't want included in the chart, add them to a `.helmignore` file in your project directory.

## Project Makefile

All projects must have a `Makefile` that implements the targets:

- `lint`:  Should do preliminary checks to confirm the project is ready for packaging.
- `package`:  Should create the chart tarball and push it up to the `helm-charts/docs` directory.
- `test`:  Typically run by Travis to test the release.

A default `common.mk` file is included that can be used for this purpose, although a project is free to implement these targets as they see fit.

## Helm-charts Makefile

The default target of the top-level Makefile packages all the charts
and then builds the `index.yaml` file based on the tarballs.

If any of the `lint`, `package`, or `test` targets are invoked, they are recursively invoked on each of the projects.  To run over a particular subset of projects just define a value on the command line.  e.g. `make package CHARTS=sample-project`

## Helm install a project:

To install the Helm chart for a particular project...

```bash
kubectl create serviceaccount --namespace kube-system tiller

kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller

helm init --service-account tiller

helm repo add lightbend-helm-charts https://repo.lightbend.com/helm-charts

helm repo update

helm install lightbend-helm-charts/<chart-name> --name=<release-name> --namespace=lightbend --debug
```

### Upgrade

```
helm repo update
helm upgrade lightbend-helm-charts/<chart-name> --name=<release-name> --namespace=lightbend --debug
```

## Cutting a Release / Publishing Charts

### Release Charts

Install [yq](https://github.com/mikefarah/yq) if you don't have it yet:

    go get github.com/mikefarah/yq
    # or
    brew install yq                  

Then run the release script on a clean master checkout:

    ./scripts/make-release.sh enterprise-suite
    git push --follow-tags
    
The `make-release.sh` script will create the chart tarball, push it to the `docs`
directory, rebuild the `index.yaml` file, and then make a
commit. Finally, `git push --tags` will publish the release and git tag.

By default, the build uses the version specified in the `Chart.yaml`
file.  The version number will be auto-incremented for the next build.
Optionally, you can specify the version to use:

    ./scripts/make-release.sh enterprise-suite 1.0.0

This is useful if you want to increment the major or minor version
number.  In the example above, the build would use v1.0.0 and `Chart.yaml` would then
be setup for the next build with version 1.0.1.

### Enterprise Suite Console

See [ES Build and Release](https://docs.google.com/document/d/14L3Zdwc-MkCDR1-7fWQYQT3k53vLc4cehAKEuOnwhxs/edit)
for details on the overall release process.  (Note that for the
`enterprise-suite` project, we automatically build and publish the
`enterprise-suite-latest` project at the same time.)

## Maintenance

Enterprise Suite Team <es-all@lightbend.com>

## License

Copyright (C) 2018 Lightbend Inc. (https://www.lightbend.com).

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this project except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
