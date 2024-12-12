{{- define "namespace-watcher-informer.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "namespace-watcher-informer.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Chart.Name .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "namespace-watcher-informer.labels" -}}
app: {{ template "namespace-watcher-informer.name" . }}
app.kubernetes.io/name: {{ template "namespace-watcher-informer.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: "watcher"
{{- end -}}

{{- define "namespace-watcher-informer.excludedNamespaces" -}}
{{- $excludedNamespaces := .Values.configs.excludedNamespaces -}}
{{- $excludedNamespaces = append $excludedNamespaces .Release.Namespace }}
{{- $excludedNamespaces | join "," -}}
{{- end -}}
