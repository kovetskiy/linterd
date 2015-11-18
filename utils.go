package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func execute(cmd *exec.Cmd) (string, error) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			if len(output) > 0 {
				return string(output), fmt.Errorf(
					"executing %s failed: %s, output: %s",
					strings.Join(cmd.Args, " "), err, output,
				)
			}

			return string(output), fmt.Errorf(
				"executing %s failed: %s, output is empty",
				strings.Join(cmd.Args, " "), err,
			)
		default:
			return string(output), fmt.Errorf(
				"executing %s failed: %s",
				strings.Join(cmd.Args, " "), err,
			)
		}
	}

	return string(output), nil
}
