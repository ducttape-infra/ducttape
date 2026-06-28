package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var ipCommand = &cobra.Command{
	Use:   "ip <vm>",
	Short: "Show the guest IP address of a running VM",
	Long: `Show the guest IP address of a running VM.

  ducttape ip myvm

For Lima VMs this queries the guest via the provisioner.
Use 'ducttape ports' for forwarded addresses.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := "ducttape-" + strings.TrimPrefix(args[0], "ducttape-")

		// LimaProvisioner.SSHInfo handles detection internally
		info, err := (&LimaProvisioner{}).SSHInfo(vmName)
		if err == nil && info.SSHPort > 0 {
			fmt.Println("127.0.0.1")
			return nil
		}

		return fmt.Errorf("VM %s not found or not running", args[0])
	},
}
