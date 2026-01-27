{{/*
Expand the name of the chart.
*/}}
{{- define "sock-shop-litmus.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "sock-shop-litmus.fullname" -}}
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
{{- define "sock-shop-litmus.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "sock-shop-litmus.labels" -}}
helm.sh/chart: {{ include "sock-shop-litmus.chart" . }}
{{ include "sock-shop-litmus.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "sock-shop-litmus.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sock-shop-litmus.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Sock-Shop namespace
*/}}
{{- define "sock-shop-litmus.sockShopNamespace" -}}
{{- .Values.namespaces.sockShop }}
{{- end }}

{{/*
Litmus namespace
*/}}
{{- define "sock-shop-litmus.litmusNamespace" -}}
{{- .Values.namespaces.litmus }}
{{- end }}

{{/*
Monitoring namespace
*/}}
{{- define "sock-shop-litmus.monitoringNamespace" -}}
{{- .Values.namespaces.monitoring }}
{{- end }}
