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

Make sure you've done any required changes to your project, including to the `values.yaml` file.

Then run the release script on a clean master checkout.  (Make sure you're in a clean master checkout and not your
working copy as stray files could end up in the tarball.  You may need to `git clone git@github.com:lightbend/helm-charts.git` the repo into a fresh directory.)

    scripts/make-release.sh enterprise-suite
    # Check that things look right at this point.  See below...
	git push --follow-tags
    
The `make-release.sh` script will create the chart tarball in the `docs`
directory, rebuild the `index.yaml` file, make
commits and generate a tag for the release. _This is the point to do a quick sanity check on things._  (For
example, check the size of, and files in, the new tarball.) If things look wrong, you can do a `git reset --hard HEAD~2` and start over.  (You'll have to delete the generated tag as well.  `git tag -d blah`)

By default, the build uses the version specified in the `Chart.yaml`
file. Optionally, you can specify the version to use:

    scripts/make-release.sh enterprise-suite 1.0.0

This is useful if you want to increment the major or minor version
number.  Either way, the patch component of the version number will be auto-incremented for the next build.
In the example above, the build would use v1.0.0 and `Chart.yaml` would then
be setup for the next build with version 1.0.1.  (The modified `Chart.yaml` is committed on its own after the commit (and tag) for the release.)

Finally, `git push --follow-tags` will push the changes (including tag) to the upstream master branch.  This will kick off a build job to publish the release to the public helm repo. You can confirm things are published with

    helm repo update
    helm search enterprise-suite

Once all that is done, announce the release.  For a minor release a
note to the #console Slack group would suffice.  A major release would
warrant an email to a wider group.

## Maintenance

Enterprise Suite Team <es-all@lightbend.com>

## License

Copyright (C) 2018 Lightbend Inc. (https://www.lightbend.com).

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this project except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
