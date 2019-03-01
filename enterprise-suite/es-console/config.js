{{- /* Concatenates the key/values of the config into a list of JSON key=value, which are finally joined with ','. */ -}}
{{- $settings := list }}
{{- range $k, $v := .Values.consoleUIConfig }}
{{- $settings = append $settings (printf "\"%s\"=\"%v\"" $k $v) }}
{{- end}}
window.installedConfig = {
  {{ $settings | join "," }}
}
