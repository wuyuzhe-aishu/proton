{{/* vim: set filetype=mustache: */}}
{{/* Expand the name of the chart. */}}
{{- define "prometheus.name" -}}
{{- if contains "prometheus" .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name "prometheus" | trunc 63 | trimSuffix "-" }}
{{- end -}}
{{- end -}}

{{/* Generate prometheus image */}}
{{- define "prometheus.image" -}}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.prometheus.repository .Values.image.prometheus.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.prometheus.repository .Values.image.prometheus.tag -}}
{{- end -}}
{{- end -}}


{{/* Generate busybox image */}}
{{- define "busybox.image" -}}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.busybox.repository .Values.image.busybox.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.busybox.repository .Values.image.busybox.tag -}}
{{- end -}}
{{- end -}}


{{/* Generate busybox image */}}
{{- define "configmapReload.image" -}}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.configmapReload.repository .Values.image.configmapReload.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.configmapReload.repository .Values.image.configmapReload.tag -}}
{{- end -}}
{{- end -}}

