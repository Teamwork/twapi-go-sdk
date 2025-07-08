//go:build !linux && !windows && !darwin && !openbsd && !freebsd && !netbsd
// +build !linux,!windows,!darwin,!openbsd,!freebsd,!netbsd

package browser

import (
	"fmt"
	"runtime"
)

func openBrowser(string, func(program string, args ...string) error) error {
	return fmt.Errorf("openBrowser: unsupported operating system: %v", runtime.GOOS)
}
