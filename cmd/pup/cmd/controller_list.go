package cmd

import (
	"fmt"

	"github.com/pupload/pupload/internal/cli/project"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all controllers in the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		controllers, err := project.ListControllers()
		if err != nil {
			return err
		}

		if len(controllers) == 0 {
			fmt.Println("No controllers configured.")
			return nil
		}

		for _, ctrl := range controllers {
			fmt.Printf("%s\t%s\n", ctrl.Name, ctrl.URL)
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(listCmd)
}
