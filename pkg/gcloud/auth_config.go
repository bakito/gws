package gcloud

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var oauthConfig = &oauth2.Config{
	ClientID:     "32555940559.apps.googleusercontent.com",
	ClientSecret: "ZmssLNjJy2998hD4CTg2ejr2",
	Scopes: []string{
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/appengine.admin",
		"https://www.googleapis.com/auth/sqlservice.login",
		"https://www.googleapis.com/auth/compute",
	},
	Endpoint: google.Endpoint,
}
