//go:build windows

package passwd

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func Prompt(prompt string) (string, error) {
	fmt.Printf("%s \n", prompt)
	key, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	return string(key), nil
}
