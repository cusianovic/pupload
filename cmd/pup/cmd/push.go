package cmd

import (
	"fmt"

	"github.com/pupload/pupload/internal/cli/project"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push <controller-name>",
	Short: "Push the project to a controller",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := project.PushProjectToController(name); err != nil {
			return err
		}

		fmt.Printf("Project pushed to controller %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
