package events

import (
	"iter"
	"strings"

	"github.com/kjanat/slimacademy/internal/models"
)

// Stream generates a loss-less sequence of events from Google Docs JSON
func Stream(book *models.Book) iter.Seq[Event] {
	return func(yield func(Event) bool) {
		// Start document
		if !yield(Event{Kind: StartDoc, Arg: book.Title}) {
			return
		}

		// Process body content
		chapterMap := buildChapterMap(book.Chapters)
		inListBlock := false

		for _, element := range book.Content.Body.Content {
			if element.Table != nil {
				// End list block if we were in one
				if inListBlock {
					if !yield(Event{Kind: EndList, Arg: nil}) {
						return
					}
					inListBlock = false
				}
				if !streamTable(element.Table, yield) {
					return
				}
			} else if element.Paragraph != nil {
				paragraph := element.Paragraph

				// Handle chapter headings
				if paragraph.ParagraphStyle.HeadingID != nil {
					if chapter, exists := chapterMap[*paragraph.ParagraphStyle.HeadingID]; exists {
						// End list block if we were in one
						if inListBlock {
							if !yield(Event{Kind: EndList, Arg: nil}) {
								return
							}
							inListBlock = false
						}
						level := getHeadingLevel(paragraph.ParagraphStyle.NamedStyleType)
						headingInfo := HeadingInfo{
							Level:    level,
							Text:     chapter.Title,
							AnchorID: slugify(chapter.Title),
						}
						if !yield(Event{Kind: StartHeading, Arg: headingInfo}) {
							return
						}
						if !yield(Event{Kind: EndHeading, Arg: nil}) {
							return
						}
						continue
					}
				}

				text := extractParagraphText(paragraph)
				hasInlineObjects := hasInlineObjects(paragraph)

				if text != "" || hasInlineObjects {
					if paragraph.ParagraphStyle.NamedStyleType == "HEADING_1" ||
						paragraph.ParagraphStyle.NamedStyleType == "HEADING_2" ||
						paragraph.ParagraphStyle.NamedStyleType == "HEADING_3" {
						// End list block if we were in one
						if inListBlock {
							if !yield(Event{Kind: EndList, Arg: nil}) {
								return
							}
							inListBlock = false
						}
						level := getHeadingLevel(paragraph.ParagraphStyle.NamedStyleType)
						headingInfo := HeadingInfo{
							Level:    level,
							Text:     text,
							AnchorID: slugify(text),
						}
						if !yield(Event{Kind: StartHeading, Arg: headingInfo}) {
							return
						}
						if !yield(Event{Kind: EndHeading, Arg: nil}) {
							return
						}
					} else if paragraph.Bullet != nil {
						// Handle bullet list items
						if !inListBlock {
							listInfo := ListInfo{
								Level:   0,
								Ordered: false,
							}
							if !yield(Event{Kind: StartList, Arg: listInfo}) {
								return
							}
							inListBlock = true
						}

						if !streamParagraph(paragraph, book, yield) {
							return
						}
					} else {
						// Regular paragraph
						if inListBlock {
							if !yield(Event{Kind: EndList, Arg: nil}) {
								return
							}
							inListBlock = false
						}

						if !yield(Event{Kind: StartParagraph, Arg: nil}) {
							return
						}
						if !streamParagraph(paragraph, book, yield) {
							return
						}
						if !yield(Event{Kind: EndParagraph, Arg: nil}) {
							return
						}
					}
				}
			}
		}

		// End list block if we're still in one at the end
		if inListBlock {
			if !yield(Event{Kind: EndList, Arg: nil}) {
				return
			}
		}

		// End document
		yield(Event{Kind: EndDoc, Arg: nil})
	}
}

// streamTable generates events for a table
func streamTable(table *models.Table, yield func(Event) bool) bool {
	if len(table.TableRows) == 0 {
		return true
	}

	tableInfo := TableInfo{
		Columns: table.Columns,
		Rows:    table.Rows,
	}
	if !yield(Event{Kind: StartTable, Arg: tableInfo}) {
		return false
	}

	for _, row := range table.TableRows {
		if !yield(Event{Kind: StartTableRow, Arg: nil}) {
			return false
		}
		for _, cell := range row.TableCells {
			if !yield(Event{Kind: StartTableCell, Arg: nil}) {
				return false
			}
			for _, element := range cell.Content {
				if element.Paragraph != nil {
					if !streamParagraph(element.Paragraph, nil, yield) {
						return false
					}
				}
			}
			if !yield(Event{Kind: EndTableCell, Arg: nil}) {
				return false
			}
		}
		if !yield(Event{Kind: EndTableRow, Arg: nil}) {
			return false
		}
	}

	return yield(Event{Kind: EndTable, Arg: nil})
}

// streamParagraph generates events for a paragraph's content
func streamParagraph(paragraph *models.Paragraph, book *models.Book, yield func(Event) bool) bool {
	var currentStyle Style
	var currentLinkURL string

	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			newStyle, newLinkURL := convertTextStyle(element.TextRun.TextStyle)

			// Generate formatting transitions
			if newStyle != currentStyle || newLinkURL != currentLinkURL {
				close, open := currentStyle.diff(newStyle)

				// Close formatting
				for _, style := range close {
					if !yield(Event{Kind: EndFormatting, Arg: FormatInfo{Style: style, URL: currentLinkURL}}) {
						return false
					}
				}

				// Open formatting
				for _, style := range open {
					if !yield(Event{Kind: StartFormatting, Arg: FormatInfo{Style: style, URL: newLinkURL}}) {
						return false
					}
				}

				currentStyle = newStyle
				currentLinkURL = newLinkURL
			}

			// Emit text content after formatting is properly set
			if element.TextRun.Content != "" {
				if !yield(Event{Kind: Text, Arg: element.TextRun.Content}) {
					return false
				}
			}
		} else if element.InlineObjectElement != nil && book != nil {
			// Handle inline images
			if book.InlineObjectMap != nil {
				if imageURL, exists := book.InlineObjectMap[element.InlineObjectElement.InlineObjectID]; exists {
					imageInfo := ImageInfo{
						URL: imageURL,
						Alt: "Image",
					}
					if !yield(Event{Kind: Image, Arg: imageInfo}) {
						return false
					}
				}
			}
		}
	}

	// Close any remaining formatting at paragraph end
	if currentStyle != 0 {
		for i := len(precedenceOrder) - 1; i >= 0; i-- {
			style := precedenceOrder[i]
			if currentStyle.Has(style) {
				if !yield(Event{Kind: EndFormatting, Arg: FormatInfo{Style: style, URL: currentLinkURL}}) {
					return false
				}
			}
		}
	}

	return true
}

// Helper functions

func buildChapterMap(chapters []models.Chapter) map[string]*models.Chapter {
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

func extractParagraphText(paragraph *models.Paragraph) string {
	var text strings.Builder
	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			text.WriteString(element.TextRun.Content)
		}
	}
	return strings.TrimSpace(text.String())
}

func hasInlineObjects(paragraph *models.Paragraph) bool {
	for _, element := range paragraph.Elements {
		if element.InlineObjectElement != nil {
			return true
		}
	}
	return false
}

func getHeadingLevel(namedStyle string) int {
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

func slugify(text string) string {
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
