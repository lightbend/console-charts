# helm-charts

## Publishing Charts

### enterprise-suite

Releasing the enterprise-suite is partially automated. To release, see *Enterprise Suite Release Process* on Google Drive.

### reactive-sandbox

> These are manual steps until we have the time to automate them.

1. First, make sure you've published the new image. See *Platform Tooling Release Process* on Google Drive.

2. Edit `reactive-sandbox/Chart.yaml`, updating version as necessary.

3. Create the package: `helm package reactive-sandbox` or `helm package enterprise-suite`

4. Move the file into docs: `mv reactive-sandbox-0.1.3.tgz docs`

5. Update the index: `helm repo index docs --url https://lightbend.github.io/helm-charts`

6. Commit and push your work

## Maintenance

Enterprise Suite Platform Team <es-platform@lightbend.com>

## License

Copyright (C) 2017 Lightbend Inc. (https://www.lightbend.com).

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this project except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
