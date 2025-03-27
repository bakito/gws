package gcloud

const (
	clientID     = "{{ .ClientID }}"
	clientSecret = "{{ .ClientSecret }}"
)

var clientScopes = []string{
{{- range .ClientScopes }}
	"{{ . }}",
{{- end }}
}
