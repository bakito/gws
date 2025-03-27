package gcloud

const (
	defaultClientID     = "{{ .DefaultClientID }}" //nolint:unused
	defaultClientSecret = "{{ .DefaultClientSecret }}" //nolint:unused
	appClientID         = "{{ .AppClientID }}" //nolint:unused
	appClientSecret     = "{{ .AppClientSecret }}" //nolint:unused
)

var (
    //nolint:unused
 	defaultClientScopes = []string{
{{- range .DefaultClientScopes }}
		"{{ . }}",
{{- end }}
	}
	//nolint:unused
	appClientScopes = []string{
 		"openid",
 		"https://www.googleapis.com/auth/userinfo.email",
 		"https://www.googleapis.com/auth/cloud-platform",
 		"https://www.googleapis.com/auth/sqlservice.login",
	}
)
