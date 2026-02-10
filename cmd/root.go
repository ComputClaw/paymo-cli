package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
)

// SetVersionInfo sets version information from build-time ldflags.
func SetVersionInfo(v, commit, date string) {
	version = v
	rootCmd.Version = v
	rootCmd.Long = fmt.Sprintf(`paymo %s — a command-line client for Paymo that allows you to:
- Track time with start/stop commands
- Manage projects and tasks
- Generate reports and timesheets
- Integrate with your development workflow

Check for updates: https://github.com/mbundgaard/paymo-cli/releases`, v)
	rootCmd.SetVersionTemplate(`{{.Name}} version {{.Version}}
`)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "paymo",
	Short: "A CLI client for Paymo time tracking and project management",
	Long: `paymo dev — a command-line client for Paymo that allows you to:
- Track time with start/stop commands
- Manage projects and tasks
- Generate reports and timesheets
- Integrate with your development workflow

Check for updates: https://github.com/mbundgaard/paymo-cli/releases`,
	Version: version,
}

// helpCmd provides help for commands (standard CLI convention)
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Help about any command",
	Long: `Help provides help for any command in the application.
Simply type paymo help [path to command] for full details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			rootCmd.Help()
			return
		}

		// Find the command
		targetCmd, _, err := rootCmd.Find(args)
		if err != nil || targetCmd == nil {
			fmt.Printf("Unknown command: %s\n", strings.Join(args, " "))
			fmt.Println("\nRun 'paymo --help' for usage.")
			return
		}

		targetCmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// GetOutputFormat returns the configured output format for use in error handling
func GetOutputFormat() string {
	f := viper.GetString("format")
	if f == "" {
		return "table"
	}
	return f
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/paymo-cli/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("format", "f", "table", "output format: table, json, csv")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "minimal output (IDs only for create/mutate commands)")
	rootCmd.PersistentFlags().Bool("no-cache", false, "bypass cache, force fresh API calls")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("no_cache", rootCmd.PersistentFlags().Lookup("no-cache"))

	// Let main.go handle error output (needed for JSON structured errors)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	// Disable Cobra's default help command (we use our own)
	rootCmd.SetHelpCommand(helpCmd)

	// Set up version template
	rootCmd.SetVersionTemplate(`{{.Name}} version {{.Version}}
`)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".paymo" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".paymo")
	}

	// Environment variables
	viper.SetEnvPrefix("PAYMO")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("verbose") {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}