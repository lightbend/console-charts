{{- define "imagePullSecret" }}
{{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .Values.imageCredentials.registry (printf "%s:%s" .Values.imageCredentials.username .Values.imageCredentials.password | b64enc) | b64enc }}
{{- end }}

# The following crazyness simply joins the given list with commas. Eg.:
#
# test:
#   - service-00:9000
#   - service-01:9000
# {{ .Values.test | include "serviceList" }}
#
# Evaluates to:
# service-00:9000,service-01:9000
{{- define "serviceList" -}}
{{- $local := dict "first" true -}}
{{- range $k, $v := . -}}{{- if not $local.first -}},{{- end -}}{{- $v -}}{{- $_ := set $local "first" false -}}{{- end -}}
{{- end -}}
