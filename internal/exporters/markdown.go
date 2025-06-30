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


	for _, element := range book.Content.Body.Content {
		if element.Table != nil {
			e.renderMarkdownTable(md, element.Table)
		} else if element.Paragraph != nil {
			paragraph := element.Paragraph

			if paragraph.ParagraphStyle.HeadingID != nil {
				if chapter, exists := chapterMap[*paragraph.ParagraphStyle.HeadingID]; exists {
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
					level := e.getHeadingLevel(paragraph.ParagraphStyle.NamedStyleType)
					md.WriteString(fmt.Sprintf("\n%s %s\n\n",
						strings.Repeat("#", level), text))
				} else if paragraph.Bullet != nil {
					// Handle bullet list items
					indent := ""
					if paragraph.Bullet.NestingLevel != nil && *paragraph.Bullet.NestingLevel > 0 {
						indent = strings.Repeat("  ", *paragraph.Bullet.NestingLevel)
					}
					formatted := e.formatTextWithBook(paragraph, book)
					if formatted != "" {
						md.WriteString(fmt.Sprintf("%s- %s\n", indent, formatted))
					}
				} else {
					formatted := e.formatTextWithBook(paragraph, book)
					if formatted != "" {
						md.WriteString(fmt.Sprintf("%s\n\n", formatted))
					}
				}
			}
		}
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

	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			content := element.TextRun.Content
			style := element.TextRun.TextStyle

			if style.Bold != nil && *style.Bold {
				content = fmt.Sprintf("**%s**", content)
			}
			if style.Italic != nil && *style.Italic {
				content = fmt.Sprintf("*%s*", content)
			}
			if style.Underline != nil && *style.Underline {
				content = fmt.Sprintf("__%s__", content)
			}
			if style.Strikethrough != nil && *style.Strikethrough {
				content = fmt.Sprintf("~~%s~~", content)
			}
			if style.Link != nil && style.Link.URL != "" {
				content = fmt.Sprintf("[%s](%s)", content, style.Link.URL)
			}

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

	result := strings.TrimSpace(text.String())
	result = strings.ReplaceAll(result, "\n", " ")

	return result
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
