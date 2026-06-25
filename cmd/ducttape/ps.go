package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// psCommand lists VMs managed by ducttape (names with ducttape- prefix).
var psCommand = &cobra.Command{
	Use:   "ps",
	Short: "List running VMs",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("VMs:")
		entries, err := os.ReadDir(configDir)
		if err != nil {
			fmt.Printf("  (error reading %s: %v)\n", configDir, err)
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			machineName := strings.TrimSuffix(e.Name(), ".json")
			if !strings.HasPrefix(machineName, "ducttape-") {
				continue
			}
			friendlyName := strings.TrimPrefix(machineName, "ducttape-")
			status := "Stopped"
			if vmIsRunning(machineName) {
				status = "Running"
			}
			fmt.Printf("  %s\t[%s]\n", friendlyName, status)
		}
	},
}
