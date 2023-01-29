---
title: "Index"
---
{{$block := .}}

# {{ $block.Name }}

- Reference: `{{ $block.Registry }}{{ $block.Name }}`

## Releases 

| Version  | Reference | 
|---|---|
{{- range $block.Versions }}
{{ if ne .Version "latest" -}}
| {{ if .HasReadme }}[{{ .Version }}](releases/{{ .Version }}){{ else }}{{ .Version }}{{ end }} | `{{ $block.Registry }}/{{ $block.Name }}:{{ .Name }}` |
{{- end }}
{{- end }}
