/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pupload/pupload/internal/cli/project"
	"github.com/pupload/pupload/internal/cli/run"
	"github.com/pupload/pupload/internal/cli/ui"
	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"
)

var inputs map[string]string

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

		tui, err := cmd.Flags().GetBool("tui")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		var g errgroup.Group
		if remote == "" {

			if tui {
				go func() {
					log.SetOutput(io.Discard)
					run.RunDevSilent(root)
				}()
			} else {
				g.Go(func() error {
					return run.RunDev(root)
				})
			}

			remote = "http://localhost:1234/"

			if err := waitForServer(remote, 10*time.Second); err != nil {
				return err
			}
		}

		run, flow, err := project.TestFlow(root, remote, flow_name)
		if err != nil {
			return err
		}

		for _, url := range run.WaitingURLs {
			name := url.Artifact.EdgeName
			path, ok := inputs[name]

			if !ok && !tui && !force {
				// TODO: cancel run
				return fmt.Errorf("missing input %s", name)
			}

			if err := uploadFile(path, url.PutURL); err != nil {
				return fmt.Errorf("error uploading input %s: %w", name, err)
			}
		}

		if tui {
			ui.TestFlowUI(*run, *flow)
			return nil
		} else {
			return g.Wait()
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	testCmd.Flags().StringP("remote", "r", "", "sets remote controller to listen on")
	testCmd.Flags().StringToStringVarP(&inputs, "input", "i", nil, "--flag <input_name>=<file_path>")
	testCmd.Flags().BoolP("force", "f", false, "forces run to execute with missing inputs")
	testCmd.Flags().Bool("tui", false, "enables TUI for monitoring task")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func uploadFile(path, url string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, f)
	if err != nil {
		return err
	}

	req.ContentLength = info.Size()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upload returned %s", resp.Status)
	}

	return nil
}

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
