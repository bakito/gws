package gcloud

const (
	defaultClientID     = "32555940559.apps.googleusercontent.com"
	defaultClientSecret = "ZmssLNjJy2998hD4CTg2ejr2"
)

var defaultClientScopes = []string{
	"openid",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/appengine.admin",
	"https://www.googleapis.com/auth/sqlservice.login",
	"https://www.googleapis.com/auth/compute",
}
