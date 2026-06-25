package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var pullCommand = &cobra.Command{
	Use:   "pull <alias>",
	Short: "Download a base image by alias and cache it locally",
	Long: `Download a base disk image from its alias URL and cache it
in the base images directory.  A date-stamped copy is also created.

  ducttape pull fedora-cloud
  ducttape pull centos-cloud`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		url, ok := knownAliases[alias]
		if !ok {
			return fmt.Errorf("unknown alias %q", alias)
		}

		fmt.Printf("Pulling %s ...\n", alias)
		path, err := downloadBaseImage(url, alias)
		if err != nil {
			return fmt.Errorf("pull failed: %w", err)
		}

		// Create a date-stamped copy
		dateStamp := time.Now().Format("20060102")
		datedPath := fmt.Sprintf("%s-%s.qcow2", path[:len(path)-len(".qcow2")], dateStamp)
		if err := copyFile(path, datedPath); err != nil {
			return fmt.Errorf("create dated copy: %w", err)
		}

		fmt.Printf("Cached: %s\n", path)
		fmt.Printf("Dated:  %s\n", datedPath)
		return nil
	},
}
