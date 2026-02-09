package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SchemaCommand describes a single CLI command for machine discovery
type SchemaCommand struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Usage       string          `json:"usage"`
	Aliases     []string        `json:"aliases,omitempty"`
	Flags       []SchemaFlag    `json:"flags,omitempty"`
	Subcommands []SchemaCommand `json:"subcommands,omitempty"`
}

// SchemaFlag describes a single flag
type SchemaFlag struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
	Usage     string `json:"usage"`
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Print machine-readable command schema (JSON)",
	Long: `Output a JSON document describing all available commands, their flags,
and usage information. Designed for AI agents and tooling integration.

Example:
  paymo schema
  paymo schema | jq '.commands[].name'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		schema := buildSchema(rootCmd)
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(schema)
	},
}

func buildSchema(cmd *cobra.Command) map[string]interface{} {
	commands := []SchemaCommand{}
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" || child.Name() == "completion" || child.Name() == "schema" {
			continue
		}
		commands = append(commands, buildCommandSchema(child))
	}
	return map[string]interface{}{
		"name":         cmd.Name(),
		"version":      version,
		"description":  cmd.Short,
		"commands":     commands,
		"global_flags": collectFlags(cmd),
	}
}

func buildCommandSchema(cmd *cobra.Command) SchemaCommand {
	sc := SchemaCommand{
		Name:        cmd.Name(),
		Description: cmd.Short,
		Usage:       cmd.UseLine(),
		Aliases:     cmd.Aliases,
		Flags:       collectFlags(cmd),
	}
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" {
			continue
		}
		sc.Subcommands = append(sc.Subcommands, buildCommandSchema(child))
	}
	return sc
}

func collectFlags(cmd *cobra.Command) []SchemaFlag {
	var flags []SchemaFlag
	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		flags = append(flags, SchemaFlag{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Type:      f.Value.Type(),
			Default:   f.DefValue,
			Usage:     f.Usage,
		})
	})
	return flags
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}
