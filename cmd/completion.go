package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for paymo.

To load completions:

Bash:
  $ source <(paymo completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ paymo completion bash > /etc/bash_completion.d/paymo
  # macOS:
  $ paymo completion bash > $(brew --prefix)/etc/bash_completion.d/paymo

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ paymo completion zsh > "${fpath[1]}/_paymo"

Fish:
  $ paymo completion fish | source
  # To load completions for each session, execute once:
  $ paymo completion fish > ~/.config/fish/completions/paymo.fish

PowerShell:
  PS> paymo completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, add output to your profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
