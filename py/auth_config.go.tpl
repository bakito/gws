package gcloud

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var oauthConfig = &oauth2.Config{
	ClientID:     "{{ .DefaultClientID }}",
	ClientSecret: "{{ .DefaultClientSecret }}",
	Scopes: []string{
{{- range .DefaultClientScopes }}
		"{{ . }}",
{{- end }}
	},
	Endpoint: google.Endpoint,
}
