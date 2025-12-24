{{/*
展開 Chart 名稱
*/}}
{{- define "images-filters.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
建立完整名稱（含 release 名稱）
*/}}
{{- define "images-filters.fullname" -}}
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
建立 Chart 標籤
*/}}
{{- define "images-filters.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
通用標籤
*/}}
{{- define "images-filters.labels" -}}
helm.sh/chart: {{ include "images-filters.chart" . }}
{{ include "images-filters.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
選擇器標籤
*/}}
{{- define "images-filters.selectorLabels" -}}
app.kubernetes.io/name: {{ include "images-filters.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
ServiceAccount 名稱
*/}}
{{- define "images-filters.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "images-filters.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
映像名稱
*/}}
{{- define "images-filters.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}
