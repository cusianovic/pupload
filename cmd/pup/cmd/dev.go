/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/pupload/pupload/internal/cli/project"
	"github.com/pupload/pupload/internal/cli/run"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Starts a local development controller and worker.",

	RunE: func(cmd *cobra.Command, args []string) error {

		root, err := project.GetProjectRoot()
		if err != nil {
			return fmt.Errorf("not inside a project")
		}

		return run.RunDev(root)
	},
}

func init() {
	rootCmd.AddCommand(devCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
