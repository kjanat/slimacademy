package writers

import (
	"fmt"
	"os"

	"github.com/kjanat/slimacademy/internal/streaming"
)

// DemoMinimalTemplate demonstrates the new minimal template system
func DemoMinimalTemplate() {
	writer := NewMinimalHTMLWriter()

	// Create sample events representing a typical academic document
	events := []streaming.Event{
		{
			Kind:               streaming.StartDoc,
			Title:              "Introduction to Computer Science",
			Description:        "Fundamental concepts and programming principles",
			BachelorYearNumber: "2024",
			AvailableDate:      "September 1, 2024",
			ExamDate:           "December 15, 2024",
			PageCount:          120,
		},
		{Kind: streaming.StartHeading, Level: 1, AnchorID: "chapter-1"},
		{Kind: streaming.Text, TextContent: "Chapter 1: Programming Fundamentals"},
		{Kind: streaming.EndHeading},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.Text, TextContent: "Programming is the process of creating instructions for computers. In this chapter, we'll explore "},
		{Kind: streaming.StartFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: "fundamental concepts"},
		{Kind: streaming.EndFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: " and learn about different "},
		{Kind: streaming.StartFormatting, Style: streaming.Italic},
		{Kind: streaming.Text, TextContent: "programming paradigms"},
		{Kind: streaming.EndFormatting, Style: streaming.Italic},
		{Kind: streaming.Text, TextContent: "."},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.StartHeading, Level: 2, AnchorID: "variables"},
		{Kind: streaming.Text, TextContent: "Variables and Data Types"},
		{Kind: streaming.EndHeading},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.Text, TextContent: "Variables are containers for storing data. For more information, visit "},
		{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "https://docs.python.org/3/tutorial/"},
		{Kind: streaming.Text, TextContent: "the official Python tutorial"},
		{Kind: streaming.EndFormatting, Style: streaming.Link},
		{Kind: streaming.Text, TextContent: "."},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.StartList},
		{Kind: streaming.StartListItem},
		{Kind: streaming.Text, TextContent: "Integer types for whole numbers"},
		{Kind: streaming.EndListItem},
		{Kind: streaming.StartListItem},
		{Kind: streaming.Text, TextContent: "Float types for decimal numbers"},
		{Kind: streaming.EndListItem},
		{Kind: streaming.StartListItem},
		{Kind: streaming.Text, TextContent: "String types for text data"},
		{Kind: streaming.EndListItem},
		{Kind: streaming.EndList},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.Image, ImageURL: "diagram-variables.png", ImageAlt: "Variable types diagram"},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.EndDoc},
	}

	// Generate HTML using the minimal template
	html, err := writer.ProcessEvents(events)
	if err != nil {
		fmt.Printf("Error generating HTML: %v\n", err)
		return
	}

	// Write to a demo file
	demoFile := "/tmp/minimal_template_demo.html"
	err = os.WriteFile(demoFile, []byte(html), 0644)
	if err != nil {
		fmt.Printf("Error writing demo file: %v\n", err)
		return
	}

	fmt.Printf("✅ Minimal template demo generated: %s\n", demoFile)
	fmt.Printf("📄 Document title: 'Introduction to Computer Science'\n")
	fmt.Printf("🎨 Clean, minimal design with:\n")
	fmt.Printf("   • Semantic HTML5 structure\n")
	fmt.Printf("   • Modern CSS with system fonts\n")
	fmt.Printf("   • Responsive design (mobile-friendly)\n")
	fmt.Printf("   • Accessibility features\n")
	fmt.Printf("   • Print-friendly styles\n")
	fmt.Printf("   • Clean metadata display\n")
	fmt.Printf("   • No complex navigation or tabs\n")
	fmt.Printf("   • Security-focused (XSS protection)\n")
	fmt.Printf("\n🚀 Open the file in a browser to see the result!\n")
}
