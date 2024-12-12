{{- define "namespace-watcher-watcher.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "namespace-watcher-watcher.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Chart.Name .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "namespace-watcher-watcher.labels" -}}
app: {{ template "namespace-watcher-watcher.name" . }}
app.kubernetes.io/name: {{ template "namespace-watcher-watcher.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: "watcher"
{{- end -}}

{{- define "namespace-watcher-watcher.excludedNamespaces" -}}
{{- $excludedNamespaces := .Values.configs.excludedNamespaces -}}
{{- $excludedNamespaces = append $excludedNamespaces .Release.Namespace }}
{{- $excludedNamespaces | join "," -}}
{{- end -}}
