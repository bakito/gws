//go:build windows

package gcloud

func openBrowser(authURL string) {
	_ = windowsCmd(authURL).Start()
}
