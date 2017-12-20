{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "reactive-sandbox.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
Note this function is slightly different than the default helm one - we (i.e. enterprise-suite) return
the release name if it is equal to the chart name, so that the name doesn't become (for example)
reactive-sandbox-reactive-sandbox
*/}}
{{- define "reactive-sandbox.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if ne .Chart.Name .Release.Name -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s" $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}