//+build darwin linux

package pkg

import "fmt"

// OpenBrowser open url with users' operating system default web browser.
// This works for MacOS.
func OpenBrowser(url string) error {
	return Run(fmt.Sprintf("open %s", url))
}
