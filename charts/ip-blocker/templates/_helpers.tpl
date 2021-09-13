{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "ip-blocker.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ip-blocker.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ip-blocker.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "ip-blocker.labels" -}}
helm.sh/chart: {{ include "ip-blocker.chart" . }}
{{ include "ip-blocker.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "ip-blocker.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ip-blocker.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "ip-blocker.serviceAccountName" -}}
{{ default (include "ip-blocker.fullname" .) .Values.serviceAccount.name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "ip-blocker.image" -}}
{{- if .Values.imageFullnameOverride }}
{{- .Values.imageFullnameOverride -}}
{{- else }}
{{-  $imageName := printf "%s/%s" .Values.image.repository .Values.image.name -}}
{{ $imageName }}:{{ .Values.image.tag | default .Chart.AppVersion }}
{{- end }}
{{- end }}
