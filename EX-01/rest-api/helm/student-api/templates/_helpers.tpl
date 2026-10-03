{{/*
Reusable pieces for the templates in this folder. A file whose name starts
with "_" renders nothing on its own - the other templates pull these in
with include.
*/}}

{{/*
The chart name, "student-api". nameOverride replaces it.
*/}}
{{- define "student-api.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
The base name of every object. Release "student-api" gives "student-api";
release "test" gives "test-student-api". Cut to 63 characters, the limit
for a Kubernetes name. fullnameOverride replaces it.
*/}}
{{- define "student-api.fullname" -}}
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
{{- define "student-api.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Labels for every object.
*/}}
{{- define "student-api.labels" -}}
helm.sh/chart: {{ include "student-api.chart" . }}
{{ include "student-api.selectorLabels" . }}
app.kubernetes.io/version: {{ include "student-api.tag" . | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
The labels the Deployment and the Service select pods by. Kept apart from
the labels above, because they must never change: Kubernetes refuses to
edit a Deployment's selector after it's created.
*/}}
{{- define "student-api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "student-api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
The image tag: image.tag if set, otherwise appVersion from Chart.yaml.
Stops the install if both are empty, instead of deploying an image with
no tag.
*/}}
{{- define "student-api.tag" -}}
{{- required "image.tag is empty and Chart.yaml has no appVersion - set one, e.g. --set image.tag=v0.4.5-6-gf42a581" (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}

{{/*
The full image reference. One definition for both containers, so the
migrations always run from the same build as the API.
*/}}
{{- define "student-api.image" -}}
{{- printf "%s:%s" (required "image.repository is required, e.g. ghcr.io/ritushinde36/student-rest-api" .Values.image.repository) (include "student-api.tag" .) }}
{{- end }}

{{/*
Where both containers get their environment variables: the ConfigMap for
plain settings, the Secret (synced from Vault) for the DSN.
*/}}
{{- define "student-api.envFrom" -}}
- configMapRef:
    name: {{ include "student-api.fullname" . }}-config
- secretRef:
    name: {{ include "student-api.fullname" . }}-secret
{{- end }}
