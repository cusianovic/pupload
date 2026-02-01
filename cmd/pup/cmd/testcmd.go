/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pupload/pupload/internal/cli/project"
	"github.com/pupload/pupload/internal/cli/run"
	"github.com/pupload/pupload/internal/cli/ui"

	"github.com/spf13/cobra"
)

func waitForServer(address string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(address)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server at %s did not become ready within %s", address, timeout)
}

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test <flowname>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		root, err := project.GetProjectRoot()
		if err != nil {
			return err
		}
		flow_name := args[0]

		remote, err := cmd.Flags().GetString("remote")
		if err != nil {
			return err
		}

		if remote == "" {
			go func() {
				log.SetOutput(io.Discard)
				run.RunDevSilent(root)
			}()

			remote = "http://localhost:1234/"

			if err := waitForServer(remote, 10*time.Second); err != nil {
				return err
			}
		}

		run, flow, err := project.TestFlow(root, remote, flow_name)
		if err != nil {
			return err
		}

		ui.TestFlowUI(*run, *flow)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	testCmd.Flags().String("remote", "", "sets remote controller to listen on")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
