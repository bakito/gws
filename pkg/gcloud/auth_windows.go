//go:build windows
// +build windows

package gcloud

func openBrowser(authURL string) {
	_ = exec.Command("cmd", "/c", "start", authURL).Start()
}
