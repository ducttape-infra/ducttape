package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var pullCommand = &cobra.Command{
	Use:   "pull <alias>",
	Short: "Download a base image by alias and cache it locally",
	Long: `Download a base disk image from its alias URL and cache it
in the base images directory.

The image is saved with a date stamp (e.g. fedora-cloud-20250627.qcow2).
A symlink from the short name (e.g. fedora-cloud.qcow2) points to it.

  ducttape pull fedora-cloud
  ducttape pull fedora-cloud:42
  ducttape pull centos-cloud`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		spec := args[0]

		// Parse alias and optional tag (e.g. "fedora-cloud:42" -> base="fedora-cloud", tag=":42")
		baseName, tag := spec, ""
		if colon := strings.LastIndex(spec, ":"); colon >= 0 && !strings.Contains(spec[colon:], "/") {
			baseName = spec[:colon]
			tag = spec[colon:]
		}

		url, ok := knownAliases[baseName]
		if !ok {
			return fmt.Errorf("unknown alias %q", baseName)
		}

		dateStamp := time.Now().Format("20060102")
		datedName := fmt.Sprintf("%s-%s", baseName, dateStamp)
		datedPath := filepath.Join(baseImagesDir, datedName+".qcow2")
		shortPath := filepath.Join(baseImagesDir, baseName+".qcow2")

		_ = os.MkdirAll(baseImagesDir, 0o755)

		if _, err := os.Stat(datedPath); err == nil {
			fmt.Printf("Already cached: %s\n", datedPath)
		} else {
			fmt.Printf("Pulling %s ...\n", spec)
			if _, err := downloadBaseImage(url+tag, datedName); err != nil {
				return fmt.Errorf("pull failed: %w", err)
			}
		}

		// Update symlink: short name -> dated file (relative symlink)
		if _, err := os.Lstat(shortPath); err == nil {
			os.Remove(shortPath)
		}
		if err := os.Symlink(datedName+".qcow2", shortPath); err != nil {
			return fmt.Errorf("create symlink: %w", err)
		}

		fmt.Printf("Cached: %s\n", datedPath)
		fmt.Printf("Linked: %s -> %s\n", shortPath, datedName+".qcow2")
		return nil
	},
}
