{{/* vim: set filetype=mustache: */}}
{{/* Expand the name of the chart. */}}
{{- define "grafana.name" -}}
{{- if contains "grafana" .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name "grafana" | trunc 63 | trimSuffix "-" }}
{{- end -}}
{{- end -}}

{{/* Return the proper grafana image name */}}
{{- define "grafana.image" -}}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.grafana.repository .Values.image.grafana.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.grafana.repository .Values.image.grafana.tag -}}
{{- end -}}
{{- end -}}
