package pkg

import "fmt"

//go:build:darwin+
// OpenBrowser .
func OpenBrowser(url string) error {
	return Run(fmt.Sprintf("open %s", url))
}

// TODO(@yeqown): OpenBrowser implementation for different platform
