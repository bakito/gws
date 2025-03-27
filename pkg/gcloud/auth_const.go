package gcloud

const (
	defaultClientID     = "32555940559.apps.googleusercontent.com"
	defaultClientSecret = "ZmssLNjJy2998hD4CTg2ejr2"
	appClientID         = "764086051850-6qr4p6gpi6hn506pt8ejuq83di341hur.apps.googleusercontent.com"
	appClientSecret     = "d-FL95Q19q7MQmFpd7hHD0Ty"
)

var (
	defaultClientScopes = []string{
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/appengine.admin",
		"https://www.googleapis.com/auth/sqlservice.login",
		"https://www.googleapis.com/auth/compute",
	}
	//nolint:unused
	appClientScopes = []string{
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/sqlservice.login",
	}
)
