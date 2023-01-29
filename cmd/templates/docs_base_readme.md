# Block Catalog

## Blocks 

| Block  | Latest Version  |  Reference |
|---|---|---|
{{- range .Blocks }}
| [{{ .Name }}]({{ .Name }}) | {{ .LatestVersion }}  | `{{ .Registry }}/{{ .Name }}:{{ .LatestVersion }}` |
{{- end }}
