package writers

import (
	"fmt"
	"html/template"
	"io"
	"net/url"
	"strings"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/streaming"
	"github.com/kjanat/slimacademy/internal/templates"
)

// init registers the HTML writer with the writer registry, providing its factory function and metadata.
func init() {
	Register("html", func(cfg *config.Config) WriterV2 {
		return &HTMLWriterV2{
			HTMLWriter: NewHTMLWriterWithConfig(cfg.HTML),
		}
	}, WriterMetadata{
		Name:        "HTML",
		Extension:   ".html",
		Description: "Clean HTML format",
		MimeType:    "text/html",
		IsBinary:    false,
	})
}

// HTMLWriter generates clean HTML from events
type HTMLWriter struct {
	config              *config.HTMLConfig
	template            *templates.MinimalTemplate
	out                 *strings.Builder
	content             *strings.Builder // Stores content separately for template
	documentData        *templates.TemplateData
	activeStyle         streaming.StyleFlags
	linkURL             string
	inList              bool
	inListItem          bool
	inTable             bool
	tableIsFirstRow     bool
	currentHeadingLevel int
	eventHandlers       map[streaming.EventKind]func(streaming.Event)
	useMinimalTemplate  bool // Toggle between minimal and complex templates
}

// NewHTMLWriter returns a new HTMLWriter instance with default HTML configuration.
func NewHTMLWriter() *HTMLWriter {
	return NewHTMLWriterWithConfig(nil)
}

// NewHTMLWriterWithConfig returns a new HTMLWriter using the provided configuration, or the default configuration if nil.
func NewHTMLWriterWithConfig(cfg *config.HTMLConfig) *HTMLWriter {
	if cfg == nil {
		cfg = config.DefaultHTMLConfig()
	}
	w := &HTMLWriter{
		config:       cfg,
		template:     templates.NewMinimalTemplate(),
		out:          &strings.Builder{},
		content:      &strings.Builder{},
		documentData: &templates.TemplateData{},
	}
	w.initEventHandlers()
	return w
}

// initEventHandlers initializes the event handler map
func (w *HTMLWriter) initEventHandlers() {
	w.eventHandlers = map[streaming.EventKind]func(streaming.Event){
		streaming.StartDoc:        w.handleStartDoc,
		streaming.EndDoc:          func(streaming.Event) { w.handleEndDoc() },
		streaming.StartParagraph:  func(streaming.Event) { w.handleStartParagraph() },
		streaming.EndParagraph:    func(streaming.Event) { w.handleEndParagraph() },
		streaming.StartHeading:    w.handleStartHeading,
		streaming.EndHeading:      func(streaming.Event) { w.handleEndHeading() },
		streaming.StartList:       func(streaming.Event) { w.handleStartList() },
		streaming.StartListItem:   func(streaming.Event) { w.handleStartListItem() },
		streaming.EndListItem:     func(streaming.Event) { w.handleEndListItem() },
		streaming.EndList:         func(streaming.Event) { w.handleEndList() },
		streaming.StartTable:      func(streaming.Event) { w.handleStartTable() },
		streaming.EndTable:        func(streaming.Event) { w.handleEndTable() },
		streaming.StartTableRow:   func(streaming.Event) { w.handleStartTableRow() },
		streaming.EndTableRow:     func(streaming.Event) { w.handleEndTableRow() },
		streaming.StartTableCell:  func(streaming.Event) { w.handleStartTableCell() },
		streaming.EndTableCell:    func(streaming.Event) { w.handleEndTableCell() },
		streaming.StartFormatting: w.handleStartFormatting,
		streaming.EndFormatting:   w.handleEndFormatting,
		streaming.Text:            w.handleText,
		streaming.Image:           w.handleImage,
	}
}

// Handle processes a single event
func (w *HTMLWriter) Handle(event streaming.Event) {
	if handler, ok := w.eventHandlers[event.Kind]; ok {
		handler(event)
	}
}

// handleStartDoc processes document start events
func (w *HTMLWriter) handleStartDoc(event streaming.Event) {
	// Store document metadata for template
	w.documentData.Title = event.Title
	w.documentData.Description = event.Description
	w.documentData.Metadata = make(map[string]string)

	// Collect metadata
	if event.BachelorYearNumber != "" {
		w.documentData.Metadata["Academic Year"] = event.BachelorYearNumber
	}
	if event.AvailableDate != "" {
		w.documentData.Metadata["Available"] = event.AvailableDate
	}
	if event.ExamDate != "" {
		w.documentData.Metadata["Exam Date"] = event.ExamDate
	}
	if event.CollegeStartYear > 0 {
		w.documentData.Metadata["College Start"] = fmt.Sprintf("%d", event.CollegeStartYear)
	}
	if event.PageCount > 0 {
		w.documentData.Metadata["Pages"] = fmt.Sprintf("%d", event.PageCount)
	}
	if len(event.Periods) > 0 {
		w.documentData.Metadata["Periods"] = strings.Join(event.Periods, ", ")
	}
	if event.ReadProgress != nil && event.PageCount > 0 {
		progress := *event.ReadProgress
		percentage := float64(progress) / float64(event.PageCount) * 100
		w.documentData.Metadata["Progress"] = fmt.Sprintf("%d/%d pages (%.1f%%)", progress, event.PageCount, percentage)
	}

	// Start the document content wrapper with semantic sections
	w.content.WriteString("<div class=\"document-body\">\n")
}

// writeDocumentHead writes the HTML document head section
func (w *HTMLWriter) writeDocumentHead(title string, event streaming.Event) {
	w.out.WriteString("<!DOCTYPE html>\n")
	w.out.WriteString("<html lang=\"en\">\n")
	w.out.WriteString("<head>\n")
	w.out.WriteString("    <meta charset=\"UTF-8\">\n")
	w.out.WriteString("    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	fmt.Fprintf(w.out, "    <title>%s</title>\n", w.escapeHTML(title))

	w.writeMetaTags(event)

	w.out.WriteString("    <style>\n")
	w.out.WriteString(w.getEnhancedCSS())
	w.out.WriteString("    </style>\n")
	w.out.WriteString("</head>\n")
	w.out.WriteString("<body>\n")
}

// writeMetaTags writes academic metadata tags
func (w *HTMLWriter) writeMetaTags(event streaming.Event) {
	if event.Description != "" {
		fmt.Fprintf(w.out, "    <meta name=\"description\" content=\"%s\">\n", w.escapeHTML(event.Description))
	}
	if event.BachelorYearNumber != "" {
		fmt.Fprintf(w.out, "    <meta name=\"academic-year\" content=\"%s\">\n", w.escapeHTML(event.BachelorYearNumber))
	}
	if len(event.Periods) > 0 {
		fmt.Fprintf(w.out, "    <meta name=\"periods\" content=\"%s\">\n", w.escapeHTML(strings.Join(event.Periods, ", ")))
	}
}

// writeAcademicHeader writes the academic header section
func (w *HTMLWriter) writeAcademicHeader(title string, event streaming.Event) {
	w.out.WriteString("    <header class=\"academic-header\">\n")
	fmt.Fprintf(w.out, "        <h1 class=\"book-title\">%s</h1>\n", w.escapeHTML(title))

	if event.Description != "" {
		fmt.Fprintf(w.out, "        <p class=\"book-description\">%s</p>\n", w.escapeHTML(event.Description))
	}

	w.out.WriteString("        <div class=\"metadata-grid\">\n")
	w.writeMetadataItems(event)
	w.writeProgressIndicator(event)
	w.out.WriteString("        </div>\n")
	w.out.WriteString("    </header>\n")
}

// writeMetadataItems writes individual metadata items
func (w *HTMLWriter) writeMetadataItems(event streaming.Event) {
	if event.BachelorYearNumber != "" {
		fmt.Fprintf(w.out, "            <div class=\"metadata-item\"><span class=\"label\">Academic Year:</span> %s</div>\n", w.escapeHTML(event.BachelorYearNumber))
	}
	if event.AvailableDate != "" {
		fmt.Fprintf(w.out, "            <div class=\"metadata-item\"><span class=\"label\">Available:</span> %s</div>\n", w.escapeHTML(event.AvailableDate))
	}
	if event.ExamDate != "" {
		fmt.Fprintf(w.out, "            <div class=\"metadata-item\"><span class=\"label\">Exam Date:</span> %s</div>\n", w.escapeHTML(event.ExamDate))
	}
	if event.CollegeStartYear > 0 {
		fmt.Fprintf(w.out, "            <div class=\"metadata-item\"><span class=\"label\">College Start:</span> %d</div>\n", event.CollegeStartYear)
	}
	if event.PageCount > 0 {
		fmt.Fprintf(w.out, "            <div class=\"metadata-item\"><span class=\"label\">Pages:</span> %d</div>\n", event.PageCount)
	}
	if len(event.Periods) > 0 {
		fmt.Fprintf(w.out, "            <div class=\"metadata-item\"><span class=\"label\">Periods:</span> %s</div>\n", w.escapeHTML(strings.Join(event.Periods, ", ")))
	}
}

// writeProgressIndicator writes the progress indicator if available
func (w *HTMLWriter) writeProgressIndicator(event streaming.Event) {
	if event.ReadProgress != nil {
		progress := *event.ReadProgress
		if event.PageCount > 0 {
			percentage := float64(progress) / float64(event.PageCount) * 100
			fmt.Fprintf(w.out, "            <div class=\"metadata-item progress-item\">\n")
			fmt.Fprintf(w.out, "                <span class=\"label\">Progress:</span> %d/%d pages (%.1f%%)\n", progress, event.PageCount, percentage)
			fmt.Fprintf(w.out, "                <div class=\"progress-bar\"><div class=\"progress-fill\" style=\"width: %.1f%%\"></div></div>\n", percentage)
			fmt.Fprintf(w.out, "            </div>\n")
		}
	}
}

// writeTableOfContents writes the table of contents if chapters are available
func (w *HTMLWriter) writeTableOfContents(chapters []models.Chapter) {
	if len(chapters) > 0 {
		w.out.WriteString("    <nav class=\"table-of-contents\">\n")
		w.out.WriteString("        <h2 class=\"toc-title\">Table of Contents</h2>\n")
		w.out.WriteString("        <ul class=\"toc-list\">\n")
		w.generateTOC(chapters, 0)
		w.out.WriteString("        </ul>\n")
		w.out.WriteString("    </nav>\n")
	}
}

// handleEndDoc processes document end events
func (w *HTMLWriter) handleEndDoc() {
	// Close document body section
	w.content.WriteString("</div>\n")

	// Use template to render final HTML
	w.documentData.Content = template.HTML(w.content.String())
	result, err := w.template.Render(*w.documentData)
	if err != nil {
		// Fallback to basic HTML if template fails
		w.out.WriteString("<!-- Template error: " + err.Error() + " -->\n")
		w.out.WriteString("<!DOCTYPE html><html><head><title>")
		w.out.WriteString(w.escapeHTML(w.documentData.Title))
		w.out.WriteString("</title></head><body>")
		w.out.WriteString(w.content.String())
		w.out.WriteString("</body></html>")
	} else {
		w.out.WriteString(result)
	}
}

// handleStartParagraph processes paragraph start events
func (w *HTMLWriter) handleStartParagraph() {
	w.closeListItemIfNeeded()
	w.content.WriteString("    <p>")
}

// handleEndParagraph processes paragraph end events
func (w *HTMLWriter) handleEndParagraph() {
	w.content.WriteString("</p>\n")
}

// handleStartHeading processes heading start events
func (w *HTMLWriter) handleStartHeading(event streaming.Event) {
	w.closeListItemIfNeeded()

	// Add semantic section wrapper for major headings
	if event.Level <= 2 {
		w.content.WriteString("    <section class=\"chapter-section\">\n")
	}

	w.currentHeadingLevel = event.Level
	fmt.Fprintf(w.content, "        <h%d id=\"%s\">", event.Level, event.AnchorID)
}

// handleEndHeading processes heading end events
func (w *HTMLWriter) handleEndHeading() {
	fmt.Fprintf(w.content, "</h%d>\n", w.currentHeadingLevel)
}

// handleStartListItem processes list item start events
func (w *HTMLWriter) handleStartListItem() {
	w.closeListItemIfNeeded()
	w.content.WriteString("        <li>")
	w.inListItem = true
}

// handleEndListItem processes list item end events
func (w *HTMLWriter) handleEndListItem() {
	w.closeListItemIfNeeded()
}

// handleStartList processes list start events
func (w *HTMLWriter) handleStartList() {
	w.inList = true
	w.content.WriteString("    <ul>\n")
}

// handleEndList processes list end events
func (w *HTMLWriter) handleEndList() {
	w.closeListItemIfNeeded()
	w.inList = false
	w.content.WriteString("    </ul>\n")
}

// handleStartTable processes table start events
func (w *HTMLWriter) handleStartTable() {
	w.inTable = true
	w.tableIsFirstRow = true
	w.content.WriteString("    <table style=\"border-collapse: collapse; width: 100%; margin: 20px 0;\">\n")
}

// handleEndTable processes table end events
func (w *HTMLWriter) handleEndTable() {
	w.inTable = false
	w.content.WriteString("    </table>\n")
}

// handleStartTableRow processes table row start events
func (w *HTMLWriter) handleStartTableRow() {
	w.content.WriteString("        <tr>\n")
}

// handleEndTableRow processes table row end events
func (w *HTMLWriter) handleEndTableRow() {
	w.content.WriteString("        </tr>\n")
	w.tableIsFirstRow = false
}

// handleStartTableCell processes table cell start events
func (w *HTMLWriter) handleStartTableCell() {
	tag := "td"
	style := "border: 1px solid #ddd; padding: 8px;"
	if w.tableIsFirstRow {
		tag = "th"
		style += " background-color: #f2f2f2; font-weight: bold;"
	}
	fmt.Fprintf(w.content, "            <%s style=\"%s\">", tag, style)
}

// handleEndTableCell processes table cell end events
func (w *HTMLWriter) handleEndTableCell() {
	tag := "td"
	if w.tableIsFirstRow {
		tag = "th"
	}
	fmt.Fprintf(w.content, "</%s>\n", tag)
}

// handleStartFormatting processes formatting start events
func (w *HTMLWriter) handleStartFormatting(event streaming.Event) {
	w.openHTMLTag(event.Style, event.LinkURL)
	w.activeStyle |= event.Style
	if event.Style&streaming.Link != 0 {
		w.linkURL = event.LinkURL
	}
}

// handleEndFormatting processes formatting end events
func (w *HTMLWriter) handleEndFormatting(event streaming.Event) {
	w.closeHTMLTag(event.Style)
	w.activeStyle &^= event.Style
	if event.Style&streaming.Link != 0 {
		w.linkURL = ""
	}
}

// handleText processes text events
func (w *HTMLWriter) handleText(event streaming.Event) {
	text := event.TextContent

	// Convert newlines to HTML breaks in specific contexts
	if w.inTable || w.shouldUseLineBreaks(text) {
		text = strings.ReplaceAll(text, "\n", "<br>")
	}

	w.content.WriteString(w.escapeHTML(text))
}

// shouldUseLineBreaks determines if newlines should become <br> tags
func (w *HTMLWriter) shouldUseLineBreaks(text string) bool {
	// Use line breaks for short lines that look like contact info, addresses, etc.
	lines := strings.Split(text, "\n")
	if len(lines) <= 1 {
		return false
	}

	// If most lines are short (under 50 chars), treat as line breaks
	shortLines := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && len(trimmed) < 50 {
			shortLines++
		}
	}

	// If majority are short lines, use line breaks
	return float64(shortLines)/float64(len(lines)) > 0.6
}

// handleImage processes image events
func (w *HTMLWriter) handleImage(event streaming.Event) {
	safeImageURL := w.sanitizeURL(event.ImageURL)
	fmt.Fprintf(w.content, "<img src=\"%s\" alt=\"%s\" style=\"max-width: 100%%; height: auto;\" />",
		template.HTMLEscapeString(safeImageURL), w.escapeHTML(event.ImageAlt))
}

// closeListItemIfNeeded closes a list item if one is currently open
func (w *HTMLWriter) closeListItemIfNeeded() {
	if w.inListItem {
		w.content.WriteString("</li>\n")
		w.inListItem = false
	}
}

// openHTMLTag opens an HTML tag based on style
func (w *HTMLWriter) openHTMLTag(style streaming.StyleFlags, linkURL string) {
	if style&streaming.Bold != 0 {
		open, _ := w.config.GetBoldTags()
		w.content.WriteString(open)
	}
	if style&streaming.Italic != 0 {
		open, _ := w.config.GetItalicTags()
		w.content.WriteString(open)
	}
	if style&streaming.Underline != 0 {
		open, _ := w.config.GetUnderlineTags()
		w.content.WriteString(open)
	}
	if style&streaming.Strike != 0 {
		open, _ := w.config.GetStrikeTags()
		w.content.WriteString(open)
	}
	if style&streaming.Highlight != 0 {
		open, _ := w.config.GetHighlightTags()
		w.content.WriteString(open)
	}
	if style&streaming.Sub != 0 {
		open, _ := w.config.GetSubscriptTags()
		w.content.WriteString(open)
	}
	if style&streaming.Sup != 0 {
		open, _ := w.config.GetSuperscriptTags()
		w.content.WriteString(open)
	}
	if style&streaming.Link != 0 {
		safeURL := w.sanitizeURL(linkURL)
		fmt.Fprintf(w.content, "<a href=\"%s\">", template.HTMLEscapeString(safeURL))
	}
}

// closeHTMLTag closes an HTML tag based on style
func (w *HTMLWriter) closeHTMLTag(style streaming.StyleFlags) {
	// Close in reverse order
	if style&streaming.Link != 0 {
		w.content.WriteString("</a>")
	}
	if style&streaming.Sup != 0 {
		_, close := w.config.GetSuperscriptTags()
		w.content.WriteString(close)
	}
	if style&streaming.Sub != 0 {
		_, close := w.config.GetSubscriptTags()
		w.content.WriteString(close)
	}
	if style&streaming.Highlight != 0 {
		_, close := w.config.GetHighlightTags()
		w.content.WriteString(close)
	}
	if style&streaming.Strike != 0 {
		_, close := w.config.GetStrikeTags()
		w.content.WriteString(close)
	}
	if style&streaming.Underline != 0 {
		_, close := w.config.GetUnderlineTags()
		w.content.WriteString(close)
	}
	if style&streaming.Italic != 0 {
		_, close := w.config.GetItalicTags()
		w.content.WriteString(close)
	}
	if style&streaming.Bold != 0 {
		_, close := w.config.GetBoldTags()
		w.content.WriteString(close)
	}
}

// escapeHTML escapes HTML special characters using html/template for enhanced security
func (w *HTMLWriter) escapeHTML(text string) string {
	return template.HTMLEscapeString(text)
}

// sanitizeURL validates and sanitizes URLs to prevent XSS attacks
func (w *HTMLWriter) sanitizeURL(urlStr string) string {
	if urlStr == "" {
		return "#"
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// If URL parsing fails, return a safe default
		return "#"
	}

	// Allow only safe schemes
	switch strings.ToLower(parsedURL.Scheme) {
	case "http", "https", "mailto", "tel":
		// These schemes are considered safe
		return parsedURL.String()
	case "":
		// Relative URLs are allowed but ensure they don't start with javascript:
		cleaned := strings.TrimSpace(urlStr)
		if strings.HasPrefix(strings.ToLower(cleaned), "javascript:") ||
			strings.HasPrefix(strings.ToLower(cleaned), "data:") ||
			strings.HasPrefix(strings.ToLower(cleaned), "vbscript:") {
			return "#"
		}
		return cleaned
	default:
		// All other schemes (javascript:, data:, vbscript:, etc.) are blocked
		return "#"
	}
}

// getCSS returns the CSS styles for the HTML document
func (w *HTMLWriter) getCSS() string {
	return `body {
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

        p {
            margin: 1em 0;
            text-align: justify;
        }

        a {
            color: #3498db;
        }

        ul {
            margin: 1em 0;
            padding-left: 2em;
        }

        table {
            margin: 1em 0;
        }

        mark {
            background-color: #fff3cd;
            padding: 0.2em;
        }
    `
}

// getEnhancedCSS returns enhanced CSS with academic styling and metadata display
func (w *HTMLWriter) getEnhancedCSS() string {
	return `@import url('https://fonts.googleapis.com/css2?family=Noto+Sans:ital,wght@0,100..900;1,100..900&family=Open+Sans:ital,wght@0,300..800;1,300..800&display=swap');

        body {
            font-family: 'Noto Sans', sans-serif;
            line-height: 1.6;
            max-width: 900px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
            background-color: #fafafa;
        }

        /* Academic Header Styling */
        .academic-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 2rem;
            margin: -20px -20px 2rem -20px;
            border-radius: 0 0 15px 15px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }

        .book-title {
            margin: 0 0 1rem 0;
            font-size: 2.5rem;
            font-weight: 300;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
        }

        .book-description {
            margin: 0 0 1.5rem 0;
            font-size: 1.1rem;
            opacity: 0.9;
            font-style: italic;
        }

        .metadata-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1rem;
            margin-top: 1.5rem;
        }

        .metadata-item {
            background: rgba(255, 255, 255, 0.1);
            padding: 0.8rem;
            border-radius: 8px;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .metadata-item .label {
            font-weight: 600;
            color: #e8f4f8;
        }

        .progress-item {
            grid-column: 1 / -1;
        }

        .progress-bar {
            background: rgba(255, 255, 255, 0.2);
            height: 8px;
            border-radius: 4px;
            margin-top: 0.5rem;
            overflow: hidden;
        }

        .progress-fill {
            background: linear-gradient(90deg, #4CAF50, #8BC34A);
            height: 100%;
            border-radius: 4px;
            transition: width 0.3s ease;
        }

        /* Table of Contents Styling */
        .table-of-contents {
            background: white;
            margin: 2rem 0;
            padding: 2rem;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            border-left: 4px solid #3498db;
        }

        .toc-title {
            color: #2c3e50;
            margin: 0 0 1.5rem 0;
            font-size: 1.5rem;
            border-bottom: 2px solid #ecf0f1;
            padding-bottom: 0.5rem;
        }

        .toc-list {
            list-style: none;
            padding: 0;
            margin: 0;
        }

        .toc-sublist {
            list-style: none;
            padding-left: 1.5rem;
            margin: 0.5rem 0;
        }

        .toc-item {
            margin: 0.5rem 0;
        }

        .toc-item.level-0 {
            font-weight: 600;
        }

        .toc-item.level-1 {
            font-weight: 500;
            opacity: 0.9;
        }

        .toc-item.level-2 {
            font-weight: 400;
            opacity: 0.8;
        }

        .toc-link {
            color: #34495e;
            text-decoration: none;
            display: block;
            padding: 0.5rem 0.8rem;
            border-radius: 5px;
            transition: all 0.3s ease;
            border-left: 3px solid transparent;
        }

        .toc-link:hover {
            background: linear-gradient(90deg, rgba(52, 152, 219, 0.1), transparent);
            border-left: 3px solid #3498db;
            transform: translateX(5px);
        }

        /* Content Styling */
        .content {
            background: white;
            padding: 2rem;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
        }

        h1, h2, h3, h4, h5, h6 {
            color: #2c3e50;
            margin-top: 2em;
            margin-bottom: 1em;
            font-weight: 600;
        }

        h2 {
            border-left: 4px solid #3498db;
            padding-left: 1rem;
            margin-left: -1rem;
            background: linear-gradient(90deg, rgba(52, 152, 219, 0.1), transparent);
            padding: 0.5rem 0 0.5rem 1rem;
            margin-left: -1rem;
        }

        h2:has(+ h2):not(:has(*)) {
          /* underline the first of two back-to-back <h2>s */
          text-decoration: underline;
        }

        h3 {
            border-left: 3px solid #9b59b6;
            padding-left: 0.8rem;
            margin-left: -0.8rem;
        }

        p {
            margin: 1em 0;
            text-align: justify;
        }

        a {
            color: #3498db;
            text-decoration: none;
            border-bottom: 1px dotted #3498db;
            transition: all 0.3s ease;
        }

        a:hover {
            color: #2980b9;
            border-bottom: 1px solid #2980b9;
        }

        ul {
            margin: 1em 0;
            padding-left: 2em;
        }

        li {
            margin: 0.5em 0;
        }

        table {
            margin: 1em 0;
            border-collapse: collapse;
            width: 100%;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            border-radius: 8px;
            overflow: hidden;
        }

        th, td {
            border: 1px solid #ddd;
            padding: 12px;
            text-align: left;
        }

        th {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            font-weight: 600;
        }

        tr:nth-child(even) {
            background-color: #f8f9fa;
        }

        tr:hover {
            background-color: #e8f4f8;
        }

        mark {
            background-color: #fff3cd;
            padding: 0.2em;
            border-radius: 3px;
        }

        strong {
            color: #2c3e50;
            font-weight: 600;
        }

        em {
            color: #34495e;
        }

        @media (max-width: 768px) {
            body {
                padding: 10px;
            }

            .academic-header {
                margin: -10px -10px 2rem -10px;
                padding: 1.5rem;
            }

            .book-title {
                font-size: 2rem;
            }

            .metadata-grid {
                grid-template-columns: 1fr;
            }

            .table-of-contents {
                padding: 1.5rem;
            }

            .toc-sublist {
                padding-left: 1rem;
            }

            .content {
                padding: 1.5rem;
            }
        }
    `
}

// generateTOC recursively generates table of contents HTML
func (w *HTMLWriter) generateTOC(chapters []models.Chapter, level int) {
	for _, chapter := range chapters {
		chapterID := w.slugify(chapter.Title)
		indent := strings.Repeat("    ", level+2)

		fmt.Fprintf(w.out, "%s<li class=\"toc-item level-%d\">\n", indent, level)

		// Add lock/free indicator
		var statusIcon string
		switch {
		case chapter.IsLocked.Bool():
			statusIcon = "🔒"
		case chapter.IsFree.Bool():
			statusIcon = "🆓"
		default:
			statusIcon = "📖"
		}

		fmt.Fprintf(w.out, "%s    <a href=\"#%s\" class=\"toc-link\">%s %s</a>\n",
			indent, chapterID, statusIcon, w.escapeHTML(chapter.Title))

		// Recursively add subchapters
		if len(chapter.SubChapters) > 0 {
			fmt.Fprintf(w.out, "%s    <ul class=\"toc-sublist\">\n", indent)
			w.generateTOC(chapter.SubChapters, level+1)
			fmt.Fprintf(w.out, "%s    </ul>\n", indent)
		}

		fmt.Fprintf(w.out, "%s</li>\n", indent)
	}
}

// slugify converts text to a URL-friendly slug
func (w *HTMLWriter) slugify(text string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// Result returns the final HTML string
func (w *HTMLWriter) Result() string {
	return w.out.String()
}

// Reset clears the writer state for reuse
func (w *HTMLWriter) Reset() {
	w.out.Reset()
	w.content.Reset()
	w.activeStyle = 0
	w.linkURL = ""
	w.inList = false
	w.inListItem = false
	w.inTable = false
	w.tableIsFirstRow = false
	w.documentData = &templates.TemplateData{}
}

// SetOutput sets the output destination (for StreamWriter interface)
func (w *HTMLWriter) SetOutput(writer io.Writer) {
	// For string-based writers, we ignore this
	// The Result() method returns the final string
}

// HTMLWriterV2 implements the WriterV2 interface
type HTMLWriterV2 struct {
	*HTMLWriter
	stats WriterStats
}

// Handle processes a single event with error handling
func (w *HTMLWriterV2) Handle(event streaming.Event) error {
	w.HTMLWriter.Handle(event)
	w.stats.EventsProcessed++

	switch event.Kind {
	case streaming.Text:
		w.stats.TextChars += len(event.TextContent)
	case streaming.Image:
		w.stats.Images++
	case streaming.StartTable:
		w.stats.Tables++
	case streaming.StartHeading:
		w.stats.Headings++
	case streaming.StartList:
		w.stats.Lists++
	}

	return nil
}

// Flush finalizes any pending operations and returns the result
func (w *HTMLWriterV2) Flush() ([]byte, error) {
	return []byte(w.Result()), nil
}

// ContentType returns the MIME type of the output
func (w *HTMLWriterV2) ContentType() string {
	return "text/html"
}

// IsText returns true since this writer outputs text-based content
func (w *HTMLWriterV2) IsText() bool {
	return true
}

// Reset clears the writer state for reuse
func (w *HTMLWriterV2) Reset() {
	w.HTMLWriter.Reset()
	w.stats = WriterStats{}
}

// Stats returns processing statistics
func (w *HTMLWriterV2) Stats() WriterStats {
	return w.stats
}
