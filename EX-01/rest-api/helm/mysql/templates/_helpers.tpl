{{/*
Reusable pieces for the templates in this folder. A file whose name starts
with "_" renders nothing on its own - the other templates pull these in
with include.
*/}}

{{/*
The chart name, "mysql". nameOverride replaces it.
*/}}
{{- define "mysql.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
The base name of every object. Release "mysql" gives "mysql"; release
"test" gives "test-mysql". Cut to 63 characters, the limit for a
Kubernetes name. fullnameOverride replaces it.
*/}}
{{- define "mysql.fullname" -}}
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
Chart name and version, for the helm.sh/chart label.
*/}}
{{- define "mysql.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Labels for every object.
*/}}
{{- define "mysql.labels" -}}
helm.sh/chart: {{ include "mysql.chart" . }}
{{ include "mysql.selectorLabels" . }}
app.kubernetes.io/version: {{ include "mysql.tag" . | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
The labels the StatefulSet and the Service select the pod by. Kept apart
from the labels above, because they must never change: Kubernetes refuses
to edit a StatefulSet's selector after it's created.
*/}}
{{- define "mysql.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mysql.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
The image tag: image.tag if set, otherwise appVersion from Chart.yaml.
Stops the install if both are empty, instead of deploying an image with
no tag.
*/}}
{{- define "mysql.tag" -}}
{{- required "image.tag is empty and Chart.yaml has no appVersion - set one, e.g. --set image.tag=8.0" (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}

{{/*
The full image reference, e.g. mysql:8.0.
*/}}
{{- define "mysql.image" -}}
{{- printf "%s:%s" (required "image.repository is required, e.g. mysql" .Values.image.repository) (include "mysql.tag" .) }}
{{- end }}
