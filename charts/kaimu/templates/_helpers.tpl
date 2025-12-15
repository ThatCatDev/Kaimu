{{/*
Expand the name of the chart.
*/}}
{{- define "kaimu.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "kaimu.fullname" -}}
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
{{- define "kaimu.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kaimu.labels" -}}
helm.sh/chart: {{ include "kaimu.chart" . }}
{{ include "kaimu.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kaimu.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kaimu.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Backend name
*/}}
{{- define "kaimu.backend.name" -}}
{{- printf "%s-backend" (include "kaimu.fullname" .) }}
{{- end }}

{{/*
Backend labels
*/}}
{{- define "kaimu.backend.labels" -}}
{{ include "kaimu.labels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Backend selector labels
*/}}
{{- define "kaimu.backend.selectorLabels" -}}
{{ include "kaimu.selectorLabels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Frontend name
*/}}
{{- define "kaimu.frontend.name" -}}
{{- printf "%s-frontend" (include "kaimu.fullname" .) }}
{{- end }}

{{/*
Frontend labels
*/}}
{{- define "kaimu.frontend.labels" -}}
{{ include "kaimu.labels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "kaimu.frontend.selectorLabels" -}}
{{ include "kaimu.selectorLabels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kaimu.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kaimu.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Backend internal URL (for frontend to connect)
*/}}
{{- define "kaimu.backend.internalUrl" -}}
{{- printf "http://%s:%d/graphql" (include "kaimu.backend.name" .) (.Values.backend.service.port | int) }}
{{- end }}

{{/*
Image pull secrets
*/}}
{{- define "kaimu.imagePullSecrets" -}}
{{- with .Values.global.imagePullSecrets }}
imagePullSecrets:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}
