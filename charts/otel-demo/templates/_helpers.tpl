{{/*
Expand the name of the chart.
*/}}
{{- define "ace-otel-demo.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "ace-otel-demo.fullname" -}}
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
{{- define "ace-otel-demo.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ace-otel-demo.labels" -}}
helm.sh/chart: {{ include "ace-otel-demo.chart" . }}
{{ include "ace-otel-demo.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ace-otel-demo.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ace-otel-demo.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
OpenTelemetry Demo namespace
*/}}
{{- define "ace-otel-demo.appNamespace" -}}
{{- .Values.namespaces.otelDemo }}
{{- end }}

{{/*
Litmus namespace
*/}}
{{- define "ace-otel-demo.litmusNamespace" -}}
{{- .Values.namespaces.litmus }}
{{- end }}

{{/*
Monitoring namespace
*/}}
{{- define "ace-otel-demo.monitoringNamespace" -}}
{{- .Values.namespaces.monitoring }}
{{- end }}
