package exporters

import (
	"fmt"
	"os"
	"strings"

	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/pkg/exporters"
)

type MarkdownExporter struct{}

func NewMarkdownExporter() exporters.Exporter {
	return &MarkdownExporter{}
}

func (e *MarkdownExporter) Export(book *models.Book, outputPath string) error {
	content := e.generateMarkdown(book)

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func (e *MarkdownExporter) GetExtension() string {
	return "md"
}

func (e *MarkdownExporter) GetName() string {
	return "Markdown"
}

func (e *MarkdownExporter) generateMarkdown(book *models.Book) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# %s\n\n", book.Title))

	if book.Description != "" {
		md.WriteString(fmt.Sprintf("%s\n\n", book.Description))
	}

	md.WriteString("## Table of Contents\n\n")
	e.generateTableOfContents(&md, book.Chapters)
	md.WriteString("\n")

	e.generateContent(&md, book)

	return md.String()
}

func (e *MarkdownExporter) generateTableOfContents(md *strings.Builder, chapters []models.Chapter) {
	for _, chapter := range chapters {
		if chapter.IsVisible == 1 {
			indent := ""
			if chapter.ParentChapterID != nil {
				indent = "  "
			}

			fmt.Fprintf(md, "%s- [%s](#%s)\n",
				indent, chapter.Title, e.slugify(chapter.Title))

			if len(chapter.SubChapters) > 0 {
				for _, subChapter := range chapter.SubChapters {
					if subChapter.IsVisible == 1 {
						md.WriteString(fmt.Sprintf("    - [%s](#%s)\n",
							subChapter.Title, e.slugify(subChapter.Title)))
					}
				}
			}
		}
	}
}

func (e *MarkdownExporter) generateContent(md *strings.Builder, book *models.Book) {
	chapterMap := e.buildChapterMap(book.Chapters)

	inListBlock := false

	for _, element := range book.Content.Body.Content {
		if element.Table != nil {
			// End list block if we were in one
			if inListBlock {
				md.WriteString("\n")
				inListBlock = false
			}
			e.renderMarkdownTable(md, element.Table)
		} else if element.Paragraph != nil {
			paragraph := element.Paragraph

			if paragraph.ParagraphStyle.HeadingID != nil {
				if chapter, exists := chapterMap[*paragraph.ParagraphStyle.HeadingID]; exists {
					// End list block if we were in one
					if inListBlock {
						md.WriteString("\n")
						inListBlock = false
					}
					level := e.getHeadingLevel(paragraph.ParagraphStyle.NamedStyleType)
					md.WriteString(fmt.Sprintf("\n%s %s\n\n",
						strings.Repeat("#", level), chapter.Title))
					continue
				}
			}

			text := e.extractParagraphText(paragraph)
			hasInlineObjects := e.hasInlineObjects(paragraph)
			
			if text != "" || hasInlineObjects {
				if paragraph.ParagraphStyle.NamedStyleType == "HEADING_1" ||
					paragraph.ParagraphStyle.NamedStyleType == "HEADING_2" ||
					paragraph.ParagraphStyle.NamedStyleType == "HEADING_3" {
					// End list block if we were in one
					if inListBlock {
						md.WriteString("\n")
						inListBlock = false
					}
					level := e.getHeadingLevel(paragraph.ParagraphStyle.NamedStyleType)
					md.WriteString(fmt.Sprintf("\n%s %s\n\n",
						strings.Repeat("#", level), text))
				} else if paragraph.Bullet != nil {
					// Handle bullet list items
					
					// Check if we're starting a new list block
					if !inListBlock {
						// The previous paragraph already added \n\n, so we just need to ensure
						// we have proper spacing. No additional newline needed.
						inListBlock = true
					}
					
					indent := ""
					if paragraph.Bullet.NestingLevel != nil && *paragraph.Bullet.NestingLevel > 0 {
						indent = strings.Repeat("  ", *paragraph.Bullet.NestingLevel)
					}
					formatted := e.formatTextWithBook(paragraph, book)
					if formatted != "" {
						md.WriteString(fmt.Sprintf("%s- %s\n", indent, formatted))
					}
				} else {
					// Regular paragraph
					
					// End list block if we were in one
					if inListBlock {
						md.WriteString("\n")
						inListBlock = false
					}
					
					formatted := e.formatTextWithBook(paragraph, book)
					if formatted != "" {
						md.WriteString(fmt.Sprintf("%s\n\n", formatted))
					}
				}
			}
		}
	}
	
	// End list block if we're still in one at the end
	if inListBlock {
		md.WriteString("\n")
	}
}

func (e *MarkdownExporter) buildChapterMap(chapters []models.Chapter) map[string]*models.Chapter {
	chapterMap := make(map[string]*models.Chapter)

	for i := range chapters {
		chapter := &chapters[i]
		chapterMap[chapter.GDocsChapterID] = chapter

		for j := range chapter.SubChapters {
			subChapter := &chapter.SubChapters[j]
			chapterMap[subChapter.GDocsChapterID] = subChapter
		}
	}

	return chapterMap
}

func (e *MarkdownExporter) extractParagraphText(paragraph *models.Paragraph) string {
	var text strings.Builder

	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			text.WriteString(element.TextRun.Content)
		}
	}

	return strings.TrimSpace(text.String())
}

func (e *MarkdownExporter) hasInlineObjects(paragraph *models.Paragraph) bool {
	for _, element := range paragraph.Elements {
		if element.InlineObjectElement != nil {
			return true
		}
	}
	return false
}

func (e *MarkdownExporter) formatText(paragraph *models.Paragraph) string {
	return e.formatTextWithBook(paragraph, nil)
}

func (e *MarkdownExporter) formatTextWithBook(paragraph *models.Paragraph, book *models.Book) string {
	var text strings.Builder

	for i, element := range paragraph.Elements {
		if element.TextRun != nil {
			content := element.TextRun.Content
			style := element.TextRun.TextStyle

			// Apply formatting with proper spacing
			content = e.applyFormattingWithSpacing(content, style, i, paragraph.Elements)

			text.WriteString(content)
		} else if element.InlineObjectElement != nil && book != nil {
			// Handle inline images
			if book.InlineObjectMap != nil {
				if imageURL, exists := book.InlineObjectMap[element.InlineObjectElement.InlineObjectID]; exists {
					text.WriteString(fmt.Sprintf("![Image](%s)", imageURL))
				}
			}
		}
	}

	result := text.String()
	result = strings.ReplaceAll(result, "\n", " ")

	return result
}

// applyFormattingWithSpacing applies markdown formatting while ensuring proper spacing
func (e *MarkdownExporter) applyFormattingWithSpacing(content string, style models.TextStyle, index int, elements []models.Element) string {
	// Check if this text run needs formatting
	needsFormatting := (style.Bold != nil && *style.Bold) ||
		(style.Italic != nil && *style.Italic) ||
		(style.Underline != nil && *style.Underline) ||
		(style.Strikethrough != nil && *style.Strikethrough)

	// If no formatting needed, handle links and return
	if !needsFormatting {
		if style.Link != nil && style.Link.URL != "" {
			return fmt.Sprintf("[%s](%s)", content, style.Link.URL)
		}
		return content
	}

	// Extract leading and trailing spaces
	leadingSpaces := ""
	trailingSpaces := ""
	trimmedContent := content

	// Extract leading spaces
	for len(trimmedContent) > 0 && trimmedContent[0] == ' ' {
		leadingSpaces += " "
		trimmedContent = trimmedContent[1:]
	}

	// Extract trailing spaces, but handle line break edge case
	for len(trimmedContent) > 0 && trimmedContent[len(trimmedContent)-1] == ' ' {
		trailingSpaces = " " + trailingSpaces
		trimmedContent = trimmedContent[:len(trimmedContent)-1]
	}

	// Apply formatting to the trimmed content
	formatted := trimmedContent
	if style.Bold != nil && *style.Bold {
		formatted = fmt.Sprintf("**%s**", formatted)
	}
	if style.Italic != nil && *style.Italic {
		formatted = fmt.Sprintf("*%s*", formatted)
	}
	if style.Underline != nil && *style.Underline {
		formatted = fmt.Sprintf("__%s__", formatted)
	}
	if style.Strikethrough != nil && *style.Strikethrough {
		formatted = fmt.Sprintf("~~%s~~", formatted)
	}
	if style.Link != nil && style.Link.URL != "" {
		formatted = fmt.Sprintf("[%s](%s)", formatted, style.Link.URL)
	}

	// Re-add spaces outside the formatting
	return leadingSpaces + formatted + trailingSpaces
}

// needsSpaceBefore determines if the current element needs a space before it
func (e *MarkdownExporter) needsSpaceBefore(content string, index int, elements []models.Element) bool {
	// If content already starts with space or newline, no need for additional space
	if strings.HasPrefix(content, " ") || strings.HasPrefix(content, "\n") {
		return false
	}

	// If this is the first element, no space needed before
	if index == 0 {
		return false
	}

	// Check the previous element
	prevElement := elements[index-1]
	if prevElement.TextRun != nil {
		prevContent := prevElement.TextRun.Content
		// If previous element ends with space or newline, no additional space needed
		if strings.HasSuffix(prevContent, " ") || strings.HasSuffix(prevContent, "\n") {
			return false
		}
		// If previous content has actual text (not just whitespace), we need a space
		if strings.TrimSpace(prevContent) != "" {
			return true
		}
	}

	return false
}

// needsSpaceAfter determines if the current element needs a space after it
func (e *MarkdownExporter) needsSpaceAfter(content string, index int, elements []models.Element) bool {
	// If content already ends with space or newline, no need for additional space
	if strings.HasSuffix(content, " ") || strings.HasSuffix(content, "\n") {
		return false
	}

	// If this is the last element, no space needed after
	if index >= len(elements)-1 {
		return false
	}

	// Check the next element
	nextElement := elements[index+1]
	if nextElement.TextRun != nil {
		nextContent := nextElement.TextRun.Content
		// If next element starts with space or newline, no additional space needed
		if strings.HasPrefix(nextContent, " ") || strings.HasPrefix(nextContent, "\n") {
			return false
		}
		// If next content has actual text (not just whitespace), we need a space
		if strings.TrimSpace(nextContent) != "" {
			return true
		}
	}

	return false
}

func (e *MarkdownExporter) getHeadingLevel(namedStyle string) int {
	switch namedStyle {
	case "HEADING_1":
		return 2
	case "HEADING_2":
		return 3
	case "HEADING_3":
		return 4
	case "HEADING_4":
		return 5
	case "HEADING_5":
		return 6
	case "HEADING_6":
		return 6
	default:
		return 2
	}
}

func (e *MarkdownExporter) slugify(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, " ", "-")
	text = strings.ReplaceAll(text, ":", "")
	text = strings.ReplaceAll(text, "/", "-")
	text = strings.ReplaceAll(text, "\\", "-")
	text = strings.ReplaceAll(text, "?", "")
	text = strings.ReplaceAll(text, "!", "")
	text = strings.ReplaceAll(text, "(", "")
	text = strings.ReplaceAll(text, ")", "")
	text = strings.ReplaceAll(text, "&", "and")

	return text
}

func (e *MarkdownExporter) renderMarkdownTable(md *strings.Builder, table *models.Table) {
	if len(table.TableRows) == 0 {
		return
	}

	md.WriteString("\n")
	
	// Render table rows
	for i, row := range table.TableRows {
		md.WriteString("|")
		for _, cell := range row.TableCells {
			cellText := ""
			for _, element := range cell.Content {
				if element.Paragraph != nil {
					cellText += e.extractParagraphText(element.Paragraph)
				}
			}
			cellText = strings.ReplaceAll(cellText, "\n", " ")
			cellText = strings.TrimSpace(cellText)
			md.WriteString(fmt.Sprintf(" %s |", cellText))
		}
		md.WriteString("\n")
		
		// Add header separator after first row
		if i == 0 && len(row.TableCells) > 0 {
			md.WriteString("|")
			for range row.TableCells {
				md.WriteString(" --- |")
			}
			md.WriteString("\n")
		}
	}
	md.WriteString("\n")
}
