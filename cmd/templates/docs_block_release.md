---
title: {{ .Version }}
---


- Reference: `{{ .Registry }}/{{ .Repository.Name }}:{{ .Version }}`
- Created: {{ .PushTime }}
- Updated: {{ .PushTime }}
- Labels:
{{ range $key, $value := .Artifact.ExtraAttrs.Config.Labels }}
    - **{{ $key }}**: {{ $value }}
{{ end }}

{{ if .Readme }}
---
{{ .Readme }}
{{ end }}