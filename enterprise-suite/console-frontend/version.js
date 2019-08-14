{
    "description": {{ .Chart.Description | quote }},
    "version": {{ .Chart.Version | quote }},
    {{- range $k, $v := .Values }}
    {{- if eq $k "imageCredentials" }}
        {{- range $k2, $v2 := . }}
        {{- if and (eq $k2 "registry") ($v2) }}
    "imageCredentials": { {{ $k2 |quote }}: {{ $v2 | quote}} },
        {{- end}}
        {{- end}}
    {{- else if ($v) }}
    {{ $k | quote }}: {{ $v | quote }},
    {{- end}}
    {{- end}}
    "name": {{ .Chart.Name | quote }}
}
