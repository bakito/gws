package log

import "fmt"

type Logger func(string)

func SetLogger(l Logger) {
	if l != nil {
		logger = l
	}
}

var logger = func(s string) {
	fmt.Print(s)
}

func Log(s string) {
	logger(s)
}

func Logf(s string, args ...any) {
	logger(fmt.Sprintf(s, args...))
}
