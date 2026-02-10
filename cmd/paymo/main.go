package main

import (
	"errors"
	"os"
	"runtime/debug"

	"github.com/ComputClaw/paymo-cli/cmd"
	"github.com/ComputClaw/paymo-cli/internal/output"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// When installed via "go install", ldflags aren't set so version stays
	// "dev". Fall back to the VCS info Go embeds automatically.
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
	cmd.SetVersionInfo(version, commit, date)
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
