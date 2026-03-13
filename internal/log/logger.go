package log

import (
	"fmt"
	"os"
)

type Logger func(string)

func SetLogger(l Logger) {
	if l != nil {
		logger = l
	}
}

var logger = Stdout

func Log(s string) {
	logger(s)
}

func Logf(s string, args ...any) {
	logger(fmt.Sprintf(s, args...))
}

var Stdout = func(s string) {
	fmt.Fprintln(os.Stdout, s)
}

var Stderr = func(s string) {
	fmt.Fprintln(os.Stderr, s)
}

var Null = func(string) {
}
