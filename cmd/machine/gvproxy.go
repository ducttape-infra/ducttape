package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

//go:embed assets/gvproxy
var gvproxyEmbedded []byte

// gvproxyPath is set by setupEnv so the gvproxy command can exec the
// embedded binary without extracting it twice.
var gvproxyPath string

// gvproxyCommand runs the embedded gvproxy binary.  setupEnv extracts it
// into CONTAINERS_HELPER_BINARY_DIR so that the macadam/podman library
// finds it automatically.
var gvproxyCommand = &cobra.Command{
	Use:   "gvproxy",
	Short: "Run the gvproxy network helper (embedded)",
	Long:  `gvproxy is a networking helper used by the macadam/podman library.  The binary is embedded in the machine executable.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if gvproxyPath == "" {
			p, err := extractEmbeddedGVProxy()
			if err != nil {
				return err
			}
			gvproxyPath = p
		}
		return syscall.Exec(gvproxyPath, append([]string{gvproxyPath}, args...), os.Environ())
	},
}

// extractEmbeddedGVProxy writes the embedded gvproxy binary to the
// machine home bin directory ($machineHome/bin/) and returns its path.
// The caller (setupEnv) sets CONTAINERS_HELPER_BINARY_DIR to that dir.
func extractEmbeddedGVProxy() (string, error) {
	binDir := filepath.Join(machineHome, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", fmt.Errorf("create machine bin dir: %w", err)
	}
	dest := filepath.Join(binDir, "gvproxy")
	if fi, err := os.Stat(dest); err == nil && fi.Mode().IsRegular() {
		// already extracted
		return dest, nil
	}
	if err := os.WriteFile(dest, gvproxyEmbedded, 0o755); err != nil {
		return "", fmt.Errorf("write gvproxy: %w", err)
	}
	return dest, nil
}
