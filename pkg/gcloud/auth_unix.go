//go:build aix || dragonfly || freebsd || (js && wasm) || linux || nacl || netbsd || openbsd || solaris

package gcloud

import "os/exec"

func openBrowser(authURL string) {
	_ = exec.Command("xdg-open", authURL).Start()
}
