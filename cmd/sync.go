package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ComputClaw/paymo-cli/internal/api"
	"github.com/ComputClaw/paymo-cli/internal/cache"
	"github.com/ComputClaw/paymo-cli/internal/config"
	"github.com/ComputClaw/paymo-cli/internal/output"
)

var validSyncTargets = []string{"all", "me", "clients", "projects", "tasks"}

// cacheTypesForTarget maps a sync target to the cache resource types that
// must be invalidated before fetching fresh data.
var cacheTypesForTarget = map[string][]string{
	"me":       {"me"},
	"clients":  {"clients"},
	"projects": {"projects", "project", "project_by_name"},
	"tasks":    {"tasks", "task", "task_by_name", "tasklists"},
}

// coreTargets are synced by default (no args) and after login.
var coreTargets = []string{"me", "clients", "projects"}

var syncCmd = &cobra.Command{
	Use:   "sync [targets...]",
	Short: "Sync Paymo data into the local cache",
	Long: `Pre-populate the local cache by fetching data from Paymo.

With no arguments, syncs core data (me, clients, projects).
Specify targets to sync specific resources.

Valid targets: all, me, clients, projects, tasks

Examples:
  paymo sync                    # Sync core data
  paymo sync all                # Sync everything
  paymo sync projects clients   # Sync specific resources
  paymo sync tasks              # Sync only tasks`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		targets, err := parseSyncTargets(args)
		if err != nil {
			return err
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		formatter := newFormatter()

		// Invalidate cache for the requested targets before fetching
		invalidateCacheForSync(targets...)

		for _, target := range targets {
			if err := syncResource(client, target, formatter); err != nil {
				return err
			}
		}

		return nil
	},
}

// parseSyncTargets validates and expands sync target arguments.
func parseSyncTargets(args []string) ([]string, error) {
	if len(args) == 0 {
		return coreTargets, nil
	}

	for _, arg := range args {
		if !isValidTarget(arg) {
			return nil, fmt.Errorf("unknown sync target %q\nValid targets: %s", arg, strings.Join(validSyncTargets, ", "))
		}
	}

	// Expand "all"
	for _, arg := range args {
		if arg == "all" {
			return []string{"me", "clients", "projects", "tasks"}, nil
		}
	}

	return args, nil
}

func isValidTarget(target string) bool {
	for _, v := range validSyncTargets {
		if v == target {
			return true
		}
	}
	return false
}

// invalidateCacheForSync opens the cache store and invalidates the resource
// types associated with the given sync targets.
func invalidateCacheForSync(targets ...string) {
	cacheDir, err := config.GetConfigDir()
	if err != nil {
		return
	}
	cachePath := filepath.Join(cacheDir, "cache.json")

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return
	}

	store, err := cache.Open(cachePath)
	if err != nil {
		return
	}
	defer store.Close()

	var types []string
	for _, t := range targets {
		types = append(types, cacheTypesForTarget[t]...)
	}
	if len(types) > 0 {
		store.InvalidateType(types...)
	}
}

// syncResource fetches a single resource type from the API and prints progress.
func syncResource(client api.PaymoAPI, target string, formatter *output.Formatter) error {
	if formatter.Format != "json" && !formatter.Quiet {
		fmt.Fprintf(formatter.Writer, "Syncing %s... ", target)
	}

	count, err := fetchResource(client, target)
	if err != nil {
		if formatter.Format != "json" && !formatter.Quiet {
			fmt.Fprintln(formatter.Writer, "failed")
		}
		return fmt.Errorf("syncing %s: %w", target, err)
	}

	if formatter.Format != "json" && !formatter.Quiet {
		fmt.Fprintf(formatter.Writer, "done (%d items)\n", count)
	}

	return nil
}

// fetchResource calls the appropriate API method for the target and returns
// the number of items fetched.
func fetchResource(client api.PaymoAPI, target string) (int, error) {
	switch target {
	case "me":
		_, err := client.GetMe()
		if err != nil {
			return 0, err
		}
		return 1, nil
	case "clients":
		clients, err := client.GetClients()
		if err != nil {
			return 0, err
		}
		return len(clients), nil
	case "projects":
		projects, err := client.GetProjects(nil)
		if err != nil {
			return 0, err
		}
		return len(projects), nil
	case "tasks":
		tasks, err := client.GetTasks(nil)
		if err != nil {
			return 0, err
		}
		return len(tasks), nil
	default:
		return 0, fmt.Errorf("unknown target: %s", target)
	}
}

// syncAfterLogin syncs core data after a successful login.
// The user is already fetched during login validation, so we seed the cache
// with it directly and only fetch clients/projects from the API.
// Errors are non-fatal — we print a warning but don't fail the login.
func syncAfterLogin(formatter *output.Formatter, user *api.User) {
	client, err := getAPIClient()
	if err != nil {
		return
	}

	if formatter.Format != "json" && !formatter.Quiet {
		fmt.Fprintln(formatter.Writer)
	}

	// Seed "me" into the cache from the already-fetched user
	seedMeCache(user)
	if formatter.Format != "json" && !formatter.Quiet {
		fmt.Fprintf(formatter.Writer, "Syncing me... done (1 items)\n")
	}

	// Sync the remaining core targets (clients, projects)
	remaining := []string{"clients", "projects"}
	invalidateCacheForSync(remaining...)

	for _, target := range remaining {
		if err := syncResource(client, target, formatter); err != nil {
			if formatter.Format != "json" && !formatter.Quiet {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
			return
		}
	}
}

// seedMeCache writes the user directly into the cache store.
func seedMeCache(user *api.User) {
	cacheDir, err := config.GetConfigDir()
	if err != nil {
		return
	}
	cachePath := filepath.Join(cacheDir, "cache.json")
	store, err := cache.Open(cachePath)
	if err != nil {
		return
	}
	defer store.Close()
	store.Set("me", "me", user)
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
