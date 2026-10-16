{{/*
Expand the name of the chart.
*/}}
{{- define "bookinfo.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "bookinfo.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bookinfo.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bookinfo.labels" -}}
helm.sh/chart: {{ include "bookinfo.chart" . }}
{{ include "bookinfo.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bookinfo.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bookinfo.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Bookinfo namespace
*/}}
{{- define "bookinfo.appNamespace" -}}
{{- .Values.namespaces.bookInfo }}
{{- end }}

{{/*
Litmus namespace
*/}}
{{- define "bookinfo.litmusNamespace" -}}
{{- .Values.namespaces.litmus }}
{{- end }}

{{/*
Monitoring namespace
*/}}
{{- define "bookinfo.monitoringNamespace" -}}
{{- .Values.namespaces.monitoring }}
{{- end }}
