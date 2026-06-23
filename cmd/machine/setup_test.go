package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestSetupEnv(t *testing.T) {
	// Create a mock command for testing
	cmd := &cobra.Command{}
	
	// Call setupEnv
	cleanup := setupEnv(cmd)
	defer cleanup()
	
	// Check that CONTAINERS_HELPER_BINARY_DIR is set
	helperDir := os.Getenv("CONTAINERS_HELPER_BINARY_DIR")
	if helperDir == "" {
		t.Fatal("CONTAINERS_HELPER_BINARY_DIR not set")
	}
	
	// Check that the directory exists
	if _, err := os.Stat(helperDir); os.IsNotExist(err) {
		t.Fatalf("Helper directory does not exist: %s", helperDir)
	}
	
	// Check that macadam binary exists in the directory
	macadamPath := filepath.Join(helperDir, "macadam")
	if _, err := os.Stat(macadamPath); os.IsNotExist(err) {
		t.Fatalf("macadam binary not found at: %s", macadamPath)
	}
	
	// Check that gvproxy binary exists in the directory
	gvproxyPath := filepath.Join(helperDir, "gvproxy")
	if _, err := os.Stat(gvproxyPath); os.IsNotExist(err) {
		t.Fatalf("gvproxy binary not found at: %s", gvproxyPath)
	}
	
	// Check that both binaries are executable
	macadamInfo, err := os.Stat(macadamPath)
	if err != nil {
		t.Fatalf("Failed to stat macadam: %v", err)
	}
	if macadamInfo.Mode()&0111 == 0 {
		t.Fatal("macadam is not executable")
	}
	
	gvproxyInfo, err := os.Stat(gvproxyPath)
	if err != nil {
		t.Fatalf("Failed to stat gvproxy: %v", err)
	}
	if gvproxyInfo.Mode()&0111 == 0 {
		t.Fatal("gvproxy is not executable")
	}
	
	fmt.Println("All tests passed!")
	fmt.Printf("Helper directory: %s\n", helperDir)
	fmt.Printf("macadam: %s (executable: %v)\n", macadamPath, macadamInfo.Mode()&0111 != 0)
	fmt.Printf("gvproxy: %s (executable: %v)\n", gvproxyPath, gvproxyInfo.Mode()&0111 != 0)
}