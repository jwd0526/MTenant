{{/*
Common Labels
*/}}
{{- define "mtenant.labels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
{{- end }}

{{/*
Postgres Labels
*/}}
{{- define "mtenant.labels.postgres" -}}
{{ include "mtenant.labels" . }}
app.kubernetes.io/name: postgres
app.kubernetes.io/component: database
{{- end }}

{{/*
NATS Labels
*/}}
{{- define "mtenant.labels.nats" -}}
{{ include "mtenant.labels" . }}
app.kubernetes.io/name: nats
app.kubernetes.io/component: messaging
{{- end }}

{{/*
Redis Labels
*/}}
{{- define "mtenant.labels.redis" -}}
{{ include "mtenant.labels" . }}
app.kubernetes.io/name: redis
app.kubernetes.io/component: cache
{{- end }}

{{/*
Service-specific labels — pass dict with "root" and "name"
*/}}
{{- define "mtenant.labels.service" -}}
{{ include "mtenant.labels" .root }}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Go service Deployment — pass dict with "root", "name", and "port"
*/}}
{{- define "mtenant.service.deployment" -}}
{{- $svc := index .root.Values .key -}}
{{- if $svc.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .name }}
  namespace: {{ .root.Values.namespace }}
  labels:
    {{- include "mtenant.labels.service" (dict "root" .root "name" .name) | nindent 4 }}
spec:
  replicas: {{ $svc.replicas }}
  selector:
    matchLabels:
      app: {{ .name }}
  template:
    metadata:
      labels:
        app: {{ .name }}
    spec:
      containers:
      - name: {{ .name }}
        image: {{ .root.Values.registry }}/{{ .name }}:{{ $svc.tag }}
        ports:
        - containerPort: {{ .port }}
        env:
        - name: PORT
          value: {{ .port | quote }}
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        - name: POSTGRES_DB
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: database
        - name: DATABASE_URL
          value: "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres-service:5432/$(POSTGRES_DB)?sslmode=disable"
        - name: NATS_URL
          value: "nats://nats-service:4222"
        - name: REDIS_URL
          value: "redis://redis-service:6379"
{{- end }}
{{- end }}

{{/*
Go service Service — pass dict with "root", "name", and "port"
*/}}
{{- define "mtenant.service.service" -}}
apiVersion: v1
kind: Service
metadata:
  name: {{ .name }}
  namespace: {{ .root.Values.namespace }}
  labels:
    {{- include "mtenant.labels.service" (dict "root" .root "name" .name) | nindent 4 }}
spec:
  type: ClusterIP
  ports:
  - port: {{ .port }}
  selector:
    app: {{ .name }}
{{- end }}
