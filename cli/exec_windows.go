// +build windows

package cli

import (
	"fmt"
)

func ExecCommand(input ExecCommandInput) error {
	return fmt.Errorf("currently exec does not work on Windows")
}
