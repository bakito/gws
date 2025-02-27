package spinner

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
)

func Start(title string) *spinner.Spinner {
	spinny := spinner.New(spinner.CharSets[38], 100*time.Millisecond, spinner.WithWriter(os.Stdout))
	spinny.Suffix = title
	spinny.Start()
	return spinny
}
