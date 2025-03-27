package gcloud

const (
	defaultClientID     = "{{ .DefaultClientID }}"
	defaultClientSecret = "{{ .DefaultClientSecret }}"
)

var defaultClientScopes = []string{
{{- range .DefaultClientScopes }}
	"{{ . }}",
{{- end }}
}
