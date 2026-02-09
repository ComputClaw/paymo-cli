package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ComputClaw/paymo-cli/internal/cache"
	"github.com/ComputClaw/paymo-cli/internal/config"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache management commands",
	Long:  `Commands for managing the local API response cache.`,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cached data",
	RunE: func(cmd *cobra.Command, args []string) error {
		cacheDir, err := config.GetConfigDir()
		if err != nil {
			return fmt.Errorf("getting config dir: %w", err)
		}
		dbPath := filepath.Join(cacheDir, "cache.json")

		// Check if cache file exists
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			formatter := newFormatter()
			return formatter.FormatSuccess("No cache to clear.", 0)
		}

		store, err := cache.Open(dbPath)
		if err != nil {
			return fmt.Errorf("opening cache: %w", err)
		}
		defer store.Close()

		if err := store.Clear(); err != nil {
			return fmt.Errorf("clearing cache: %w", err)
		}

		formatter := newFormatter()
		return formatter.FormatSuccess("Cache cleared.", 0)
	},
}

var cacheStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cache statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		cacheDir, err := config.GetConfigDir()
		if err != nil {
			return fmt.Errorf("getting config dir: %w", err)
		}
		dbPath := filepath.Join(cacheDir, "cache.json")

		formatter := newFormatter()

		// Check if cache file exists
		info, statErr := os.Stat(dbPath)
		if os.IsNotExist(statErr) {
			if formatter.Format == "json" {
				return formatter.FormatTimerStatus(map[string]interface{}{
					"enabled":  true,
					"entries":  0,
					"size_kb":  0,
					"db_path":  dbPath,
				})
			}
			fmt.Fprintln(formatter.Writer, "Cache is empty (no database file).")
			return nil
		}

		store, err := cache.Open(dbPath)
		if err != nil {
			return fmt.Errorf("opening cache: %w", err)
		}
		defer store.Close()

		stats, err := store.Stats()
		if err != nil {
			return fmt.Errorf("reading cache stats: %w", err)
		}

		total := 0
		for _, count := range stats {
			total += count
		}

		sizeKB := int64(0)
		if info != nil {
			sizeKB = info.Size() / 1024
		}

		if formatter.Format == "json" {
			return formatter.FormatTimerStatus(map[string]interface{}{
				"enabled":    true,
				"entries":    total,
				"size_kb":    sizeKB,
				"db_path":    dbPath,
				"by_type":    stats,
			})
		}

		fmt.Fprintf(formatter.Writer, "Cache Status\n")
		fmt.Fprintf(formatter.Writer, "  Path:    %s\n", dbPath)
		fmt.Fprintf(formatter.Writer, "  Size:    %d KB\n", sizeKB)
		fmt.Fprintf(formatter.Writer, "  Entries: %d\n", total)
		if len(stats) > 0 {
			fmt.Fprintf(formatter.Writer, "  By type:\n")
			for rt, count := range stats {
				fmt.Fprintf(formatter.Writer, "    %-20s %d\n", rt, count)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheStatusCmd)
}
