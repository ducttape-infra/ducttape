package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// imagesCommand lists base and built images
var imagesCommand = &cobra.Command{
	Use:   "images",
	Short: "List base and built images",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Base images:")
		files, err := os.ReadDir(baseImagesDir)
		if err != nil {
			fmt.Printf("  (error reading %s: %v)\n", baseImagesDir, err)
		} else {
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".qcow2") {
					tag := strings.TrimSuffix(f.Name(), ".qcow2")
					fmt.Printf("  %s\n", tag)
				}
			}
		}
		fmt.Println("Built images:")
		files, err = os.ReadDir(imagesDir)
		if err != nil {
			fmt.Printf("  (error reading %s: %v)\n", imagesDir, err)
		} else {
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".qcow2") {
					tag := strings.TrimSuffix(f.Name(), ".qcow2")
					fmt.Printf("  %s\n", tag)
				}
			}
		}
	},
}
