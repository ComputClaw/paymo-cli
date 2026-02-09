package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"

	"github.com/ComputClaw/paymo-cli/internal/api"
	"github.com/ComputClaw/paymo-cli/internal/output"
)

// newFormatter creates a formatter from the current viper config
func newFormatter() *output.Formatter {
	f := output.NewFormatter(viper.GetString("format"))
	f.Quiet = viper.GetBool("quiet")
	return f
}

// resolveProjectID resolves a project argument (ID or name) to a numeric ID
func resolveProjectID(client api.PaymoAPI, arg string) (int, error) {
	if id, err := strconv.Atoi(arg); err == nil {
		return id, nil
	}
	project, err := client.GetProjectByName(arg)
	if err != nil {
		return 0, fmt.Errorf("project not found: %w", err)
	}
	return project.ID, nil
}

// resolveProject resolves a project argument (ID or name) to a full Project
func resolveProject(client api.PaymoAPI, arg string) (*api.Project, error) {
	if id, err := strconv.Atoi(arg); err == nil {
		project, err := client.GetProject(id)
		if err != nil {
			return nil, fmt.Errorf("project not found: %w", err)
		}
		return project, nil
	}
	project, err := client.GetProjectByName(arg)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}
	return project, nil
}

// resolveTask resolves a task argument (ID or name) to a full Task.
// Name-based lookup requires a project context.
func resolveTask(client api.PaymoAPI, arg string, projectFlag string) (*api.Task, error) {
	if id, err := strconv.Atoi(arg); err == nil {
		task, err := client.GetTask(id)
		if err != nil {
			return nil, fmt.Errorf("task not found: %w", err)
		}
		return task, nil
	}
	// Name-based lookup
	if projectFlag == "" {
		return nil, fmt.Errorf("task name lookup requires --project flag (or use numeric task ID)")
	}
	projectID, err := resolveProjectID(client, projectFlag)
	if err != nil {
		return nil, err
	}
	task, err := client.GetTaskByName(projectID, arg)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return task, nil
}
