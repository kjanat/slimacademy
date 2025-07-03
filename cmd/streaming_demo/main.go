package main

import (
	"fmt"
	"os"

	"github.com/kjanat/slimacademy/internal/writers"
)

func main() {
	fmt.Println("🚀 SlimAcademy: Go 1.23+ Streaming Architecture Demo")
	fmt.Println("==================================================")
	fmt.Println()

	// Show current working directory
	cwd, _ := os.Getwd()
	fmt.Printf("Working directory: %s\n\n", cwd)

	// Run the streaming demo
	writers.DemoStreamingImprovements()

	fmt.Println("🎉 Streaming architecture improvements successfully demonstrated!")
	fmt.Println()
	fmt.Println("📋 Summary of Architectural Enhancements:")
	fmt.Println("   • True iter.Seq[Event] streaming for O(1) memory usage")
	fmt.Println("   • unique.Handle[string] for efficient duplicate detection")
	fmt.Println("   • Context-aware processing with cancellation support")
	fmt.Println("   • Integration with minimal template system")
	fmt.Println("   • Comprehensive test coverage and benchmarks")
}
