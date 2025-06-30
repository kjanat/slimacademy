package exporters

import (
	"fmt"
	"os"
	"strings"

	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/pkg/exporters"
)

type HTMLExporter struct{}

func NewHTMLExporter() exporters.Exporter {
	return &HTMLExporter{}
}

func (e *HTMLExporter) Export(book *models.Book, outputPath string) error {
	content := e.generateHTML(book)

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func (e *HTMLExporter) GetExtension() string {
	return "html"
}

func (e *HTMLExporter) GetName() string {
	return "HTML"
}

func (e *HTMLExporter) generateHTML(book *models.Book) string {
	var html strings.Builder

	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html lang=\"en\">\n")
	html.WriteString("<head>\n")
	html.WriteString("    <meta charset=\"UTF-8\">\n")
	html.WriteString("    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	html.WriteString(fmt.Sprintf("    <title>%s</title>\n", e.escapeHTML(book.Title)))
	html.WriteString("    <style>\n")
	html.WriteString(e.getCSS())
	html.WriteString("    </style>\n")
	html.WriteString("</head>\n")
	html.WriteString("<body>\n")

	html.WriteString(fmt.Sprintf("    <h1>%s</h1>\n", e.escapeHTML(book.Title)))

	if book.Description != "" {
		html.WriteString(fmt.Sprintf("    <p class=\"description\">%s</p>\n", e.escapeHTML(book.Description)))
	}

	html.WriteString("    <h2>Table of Contents</h2>\n")
	html.WriteString("    <nav class=\"toc\">\n")
	e.generateHTMLTableOfContents(&html, book.Chapters)
	html.WriteString("    </nav>\n")

	html.WriteString("    <main>\n")
	e.generateHTMLContent(&html, book)
	html.WriteString("    </main>\n")

	html.WriteString("</body>\n")
	html.WriteString("</html>\n")

	return html.String()
}

func (e *HTMLExporter) getCSS() string {
	return `
        body {
            font-family: 'Georgia', serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
        }
        
        h1, h2, h3, h4, h5, h6 {
            color: #2c3e50;
            margin-top: 2em;
            margin-bottom: 1em;
        }
        
        h1 {
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        
        .description {
            font-style: italic;
            color: #666;
            font-size: 1.1em;
            margin-bottom: 2em;
        }
        
        .toc {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 5px;
            margin: 20px 0;
        }
        
        .toc ul {
            list-style-type: none;
            padding-left: 0;
        }
        
        .toc li {
            margin: 5px 0;
        }
        
        .toc a {
            text-decoration: none;
            color: #3498db;
        }
        
        .toc a:hover {
            text-decoration: underline;
        }
        
        p {
            margin: 1em 0;
            text-align: justify;
        }
        
        .bold {
            font-weight: bold;
        }
        
        .italic {
            font-style: italic;
        }
        
        .underline {
            text-decoration: underline;
        }
        
        a {
            color: #3498db;
        }
        
        main {
            margin-top: 2em;
        }
    `
}

func (e *HTMLExporter) generateHTMLTableOfContents(html *strings.Builder, chapters []models.Chapter) {
	html.WriteString("        <ul>\n")

	for _, chapter := range chapters {
		if chapter.IsVisible == 1 {
			anchor := e.slugify(chapter.Title)
			fmt.Fprintf(html, "            <li><a href=\"#%s\">%s</a></li>\n",
				anchor, e.escapeHTML(chapter.Title))

			if len(chapter.SubChapters) > 0 {
				html.WriteString("                <ul>\n")
				for _, subChapter := range chapter.SubChapters {
					if subChapter.IsVisible == 1 {
						subAnchor := e.slugify(subChapter.Title)
						fmt.Fprintf(html, "                    <li><a href=\"#%s\">%s</a></li>\n",
							subAnchor, e.escapeHTML(subChapter.Title))
					}
				}
				html.WriteString("                </ul>\n")
			}
		}
	}

	html.WriteString("        </ul>\n")
}

func (e *HTMLExporter) generateHTMLContent(html *strings.Builder, book *models.Book) {
	chapterMap := e.buildChapterMap(book.Chapters)

	for _, element := range book.Content.Body.Content {
		if element.Paragraph != nil {
			paragraph := element.Paragraph

			if paragraph.ParagraphStyle.HeadingID != nil {
				if chapter, exists := chapterMap[*paragraph.ParagraphStyle.HeadingID]; exists {
					level := e.getHeadingLevel(paragraph.ParagraphStyle.NamedStyleType)
					anchor := e.slugify(chapter.Title)
					fmt.Fprintf(html, "        <h%d id=\"%s\">%s</h%d>\n",
						level, anchor, e.escapeHTML(chapter.Title), level)
					continue
				}
			}

			text := e.extractParagraphText(paragraph)
			if text != "" {
				if paragraph.ParagraphStyle.NamedStyleType == "HEADING_1" ||
					paragraph.ParagraphStyle.NamedStyleType == "HEADING_2" ||
					paragraph.ParagraphStyle.NamedStyleType == "HEADING_3" {
					level := e.getHeadingLevel(paragraph.ParagraphStyle.NamedStyleType)
					anchor := e.slugify(text)
					fmt.Fprintf(html, "        <h%d id=\"%s\">%s</h%d>\n",
						level, anchor, e.escapeHTML(text), level)
				} else {
					formatted := e.formatHTMLText(paragraph)
					if formatted != "" {
						fmt.Fprintf(html, "        <p>%s</p>\n", formatted)
					}
				}
			}
		}
	}
}

func (e *HTMLExporter) buildChapterMap(chapters []models.Chapter) map[string]*models.Chapter {
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

func (e *HTMLExporter) extractParagraphText(paragraph *models.Paragraph) string {
	var text strings.Builder

	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			text.WriteString(element.TextRun.Content)
		}
	}

	return strings.TrimSpace(text.String())
}

func (e *HTMLExporter) formatHTMLText(paragraph *models.Paragraph) string {
	var html strings.Builder

	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			content := e.escapeHTML(element.TextRun.Content)
			style := element.TextRun.TextStyle

			if style.Bold != nil && *style.Bold {
				content = fmt.Sprintf("<strong>%s</strong>", content)
			}
			if style.Italic != nil && *style.Italic {
				content = fmt.Sprintf("<em>%s</em>", content)
			}
			if style.Underline != nil && *style.Underline {
				content = fmt.Sprintf("<u>%s</u>", content)
			}
			if style.Link != nil && style.Link.URL != "" {
				content = fmt.Sprintf("<a href=\"%s\">%s</a>",
					e.escapeHTML(style.Link.URL), content)
			}

			html.WriteString(content)
		}
	}

	result := strings.TrimSpace(html.String())
	result = strings.ReplaceAll(result, "\n", " ")

	return result
}

func (e *HTMLExporter) getHeadingLevel(namedStyle string) int {
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

func (e *HTMLExporter) slugify(text string) string {
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

func (e *HTMLExporter) escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")

	return text
}
