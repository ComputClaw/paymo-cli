package main

import (
	"errors"
	"os"

	"github.com/ComputClaw/paymo-cli/cmd"
	"github.com/ComputClaw/paymo-cli/internal/output"
)

func main() {
	if err := cmd.Execute(); err != nil {
		format := cmd.GetOutputFormat()
		formatter := output.NewFormatter(format)
		formatter.FormatError(err)

		var ec interface{ ExitCode() int }
		if errors.As(err, &ec) {
			os.Exit(ec.ExitCode())
		}
		os.Exit(1)
	}
}
