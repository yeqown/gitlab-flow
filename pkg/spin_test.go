package pkg

import (
	"testing"
	"time"
)

func Test_Spinner(t *testing.T) {
	spinner := NewSpinner(Spinner01)
	spinner.Start()
	defer spinner.Stop()

	time.Sleep(5 * time.Second)
}
