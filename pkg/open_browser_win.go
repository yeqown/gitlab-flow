//+build windows

package pkg

import "fmt"

// OpenBrowser open url with users' operating system default web browser.
// This works for windows.
// https://stackoverflow.com/questions/3739327/launching-a-website-via-windows-commandline
func OpenBrowser(url string) error {
	return Run(fmt.Sprintf("start /max %s", url))
}
