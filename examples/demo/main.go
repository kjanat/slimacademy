package main

import (
	"fmt"
	"os"

	"github.com/kjanat/slimacademy/internal/writers"
)

func main() {
	fmt.Println("🎨 SlimAcademy Minimal Template Demo")
	fmt.Println("====================================")
	fmt.Println()

	// Show current working directory
	cwd, _ := os.Getwd()
	fmt.Printf("Working directory: %s\n\n", cwd)

	// Run the demo
	writers.DemoMinimalTemplate()
}
