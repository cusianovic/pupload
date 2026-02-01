package cmd

import (
	"fmt"

	"github.com/pupload/pupload/internal/cli/project"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a controller to the project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		url := args[1]

		if err := project.AddController(name, url); err != nil {
			return err
		}

		fmt.Printf("Controller %q added (%s)\n", name, url)
		return nil
	},
}

func init() {
	controllerCmd.AddCommand(addCmd)
}
