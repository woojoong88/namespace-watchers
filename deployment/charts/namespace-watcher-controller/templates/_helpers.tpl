{{- define "namespace-watcher-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "namespace-watcher-controller.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Chart.Name .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "namespace-watcher-controller.labels" -}}
app: {{ template "namespace-watcher-controller.name" . }}
app.kubernetes.io/name: {{ template "namespace-watcher-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: "controller"
{{- end -}}

{{- define "namespace-watcher-controller.excludedNamespaces" -}}
{{- $excludedNamespaces := .Values.configs.excludedNamespaces -}}
{{- $excludedNamespaces = append $excludedNamespaces .Release.Namespace }}
{{- $excludedNamespaces | join "," -}}
{{- end -}}
