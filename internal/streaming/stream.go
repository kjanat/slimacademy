// Package streaming provides memory-efficient event-driven document processing for SlimAcademy.
// It implements Go 1.23+ iterators with O(1) memory usage, unique string interning,
// and context-aware streaming for processing large documents with minimal resource consumption.
package streaming

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"strings"
	"unicode"
	"unique"

	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/sanitizer"
)

// EventKind represents the type of event in the document stream
type EventKind uint8

const (
	StartDoc EventKind = iota
	EndDoc
	StartParagraph
	EndParagraph
	StartHeading
	EndHeading
	StartList
	EndList
	StartTable
	EndTable
	StartTableRow
	EndTableRow
	StartTableCell
	EndTableCell
	StartFormatting
	EndFormatting
	Text
	Image
)

// StyleFlags represents active formatting as bit flags
type StyleFlags uint16

const (
	Bold StyleFlags = 1 << iota
	Italic
	Underline
	Strike
	Highlight
	Sub
	Sup
	Link
)

// Event represents a single document event with concrete data
type Event struct {
	Kind EventKind

	// Document
	Title string

	// Document Metadata (for StartDoc events)
	Description        string
	AvailableDate      string
	ExamDate           string
	BachelorYearNumber string
	CollegeStartYear   int64
	ReadProgress       *int64
	ReadPercentage     any
	PageCount          int64
	HasFreeChapters    int64
	Periods            []string
	Images             []string         // URLs of all images
	Chapters           []models.Chapter // Chapter hierarchy for TOC

	// Heading
	Level       int
	HeadingText unique.Handle[string] // Interned for O(1) duplicate detection
	AnchorID    string

	// List
	ListLevel   int
	ListOrdered bool

	// Table
	TableColumns int
	TableRows    int

	// Formatting
	Style   StyleFlags
	LinkURL string

	// Content
	TextContent string
	ImageURL    string
	ImageAlt    string
}

// StreamOptions configures event streaming behavior
type StreamOptions struct {
	ChunkSize    int  // For chunking huge paragraphs with bytes.Lines
	MemoryLimit  int  // Maximum memory usage in bytes
	SkipEmpty    bool // Skip empty content
	SanitizeText bool // Apply text sanitization
}

// DefaultStreamOptions returns a StreamOptions struct with recommended default settings for chunk size, memory limit, skipping empty content, and text sanitization.
func DefaultStreamOptions() StreamOptions {
	return StreamOptions{
		ChunkSize:    1024,
		MemoryLimit:  100 * 1024 * 1024, // 100MB
		SkipEmpty:    true,
		SanitizeText: true,
	}
}

// Streamer generates events from sanitized document content
type Streamer struct {
	options   StreamOptions
	sanitizer *sanitizer.Sanitizer
	slugCache map[string]int // For duplicate slug detection
}

// NewStreamer returns a new Streamer configured with the provided streaming options.
func NewStreamer(opts StreamOptions) *Streamer {
	return &Streamer{
		options:   opts,
		sanitizer: sanitizer.NewSanitizer(),
		slugCache: make(map[string]int),
	}
}

// Stream generates a memory-efficient event sequence from a book
func (s *Streamer) Stream(ctx context.Context, book *models.Book) iter.Seq[Event] {
	return func(yield func(Event) bool) {
		// Sanitize input first
		var sanitizedBook *models.Book
		if s.options.SanitizeText {
			result := s.sanitizer.Sanitize(book)
			sanitizedBook = result.Book
		} else {
			sanitizedBook = book
		}

		// Collect image URLs
		var imageURLs []string
		for _, img := range sanitizedBook.Images {
			imageURLs = append(imageURLs, img.ImageURL)
		}

		// Start document
		if !s.yieldEvent(ctx, yield, Event{
			Kind:               StartDoc,
			Title:              sanitizedBook.Title,
			Description:        sanitizedBook.Description,
			AvailableDate:      sanitizedBook.AvailableDate,
			ExamDate:           sanitizedBook.ExamDate,
			BachelorYearNumber: sanitizedBook.BachelorYearNumber,
			CollegeStartYear:   sanitizedBook.CollegeStartYear,
			ReadProgress:       sanitizedBook.ReadProgress,
			ReadPercentage:     sanitizedBook.ReadPercentage,
			PageCount:          sanitizedBook.PageCount,
			HasFreeChapters:    sanitizedBook.HasFreeChapters.Int64(),
			Periods:            sanitizedBook.Periods,
			Images:             imageURLs,
			Chapters:           sanitizedBook.Chapters,
		}) {
			return
		}

		// Process content with memory management
		s.processContent(ctx, sanitizedBook, yield)

		// End document
		s.yieldEvent(ctx, yield, Event{Kind: EndDoc})
	}
}

// processContent handles the main document content
func (s *Streamer) processContent(ctx context.Context, book *models.Book, yield func(Event) bool) {
	chapterMap := s.buildChapterMap(book.Chapters)
	inListBlock := false

	// Handle different content types
	var content []models.StructuralElement
	if book.Content != nil {
		if book.Content.Document != nil {
			content = book.Content.Document.Body.Content
		} else if book.Content.Chapters != nil {
			// For chapter-based content, create synthetic paragraphs
			s.processChapters(ctx, book.Content.Chapters, yield)
			return
		}
	}

	for i, element := range content {
		// Check context cancellation periodically
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		if element.Table != nil {
			if inListBlock {
				if !s.yieldEvent(ctx, yield, Event{Kind: EndList}) {
					return
				}
				inListBlock = false
			}
			if !s.processTable(ctx, element.Table, yield) {
				return
			}
		} else if element.Paragraph != nil {
			if !s.processParagraph(ctx, element.Paragraph, book, chapterMap, &inListBlock, yield) {
				return
			}
		}
	}

	// End any remaining list block
	if inListBlock {
		s.yieldEvent(ctx, yield, Event{Kind: EndList})
	}
}

// processParagraph handles paragraph elements with chunking for large content
func (s *Streamer) processParagraph(ctx context.Context, paragraph *models.Paragraph, book *models.Book, chapterMap map[string]*models.Chapter, inListBlock *bool, yield func(Event) bool) bool {
	// Handle chapter headings
	if paragraph.ParagraphStyle.HeadingID != nil {
		if chapter, exists := chapterMap[*paragraph.ParagraphStyle.HeadingID]; exists {
			return s.processChapterHeading(ctx, chapter, inListBlock, yield)
		}
	}

	// Extract and validate content
	text := s.extractParagraphText(paragraph)
	hasInlineObjects := s.hasInlineObjects(paragraph)

	if s.options.SkipEmpty && text == "" && !hasInlineObjects {
		return true
	}

	// Handle different paragraph types
	switch {
	case s.isHeading(paragraph):
		return s.processHeading(ctx, paragraph, text, inListBlock, yield)
	case paragraph.Bullet != nil:
		return s.processListItem(ctx, paragraph, book, inListBlock, yield)
	default:
		return s.processRegularParagraph(ctx, paragraph, book, inListBlock, yield)
	}
}

// processChapterHeading handles chapter-based headings
func (s *Streamer) processChapterHeading(ctx context.Context, chapter *models.Chapter, inListBlock *bool, yield func(Event) bool) bool {
	trimmedTitle := strings.TrimSpace(chapter.Title)
	if s.options.SkipEmpty && trimmedTitle == "" {
		return true
	}

	if *inListBlock {
		if !s.yieldEvent(ctx, yield, Event{Kind: EndList}) {
			return false
		}
		*inListBlock = false
	}

	return s.yieldHeading(ctx, 2, trimmedTitle, yield)
}

// processHeading handles regular headings
func (s *Streamer) processHeading(ctx context.Context, paragraph *models.Paragraph, text string, inListBlock *bool, yield func(Event) bool) bool {
	trimmedText := strings.TrimSpace(text)
	if s.options.SkipEmpty && trimmedText == "" {
		return true
	}

	if *inListBlock {
		if !s.yieldEvent(ctx, yield, Event{Kind: EndList}) {
			return false
		}
		*inListBlock = false
	}

	level := s.getHeadingLevel(paragraph.ParagraphStyle.NamedStyleType)
	return s.yieldHeading(ctx, level, trimmedText, yield)
}

// yieldHeading emits heading events with unique slug generation
func (s *Streamer) yieldHeading(ctx context.Context, level int, text string, yield func(Event) bool) bool {
	anchorID := s.generateUniqueSlug(text)

	events := []Event{
		{
			Kind:        StartHeading,
			Level:       level,
			HeadingText: unique.Make(text),
			AnchorID:    anchorID,
		},
		{
			Kind:        Text,
			TextContent: text,
		},
		{Kind: EndHeading},
	}

	for _, event := range events {
		if !s.yieldEvent(ctx, yield, event) {
			return false
		}
	}
	return true
}

// processListItem handles bullet list items
func (s *Streamer) processListItem(ctx context.Context, paragraph *models.Paragraph, book *models.Book, inListBlock *bool, yield func(Event) bool) bool {
	if !*inListBlock {
		if !s.yieldEvent(ctx, yield, Event{
			Kind:        StartList,
			ListLevel:   0,
			ListOrdered: false,
		}) {
			return false
		}
		*inListBlock = true
	}

	return s.processParagraphContent(ctx, paragraph, book, yield)
}

// processRegularParagraph handles standard paragraphs
func (s *Streamer) processRegularParagraph(ctx context.Context, paragraph *models.Paragraph, book *models.Book, inListBlock *bool, yield func(Event) bool) bool {
	if *inListBlock {
		if !s.yieldEvent(ctx, yield, Event{Kind: EndList}) {
			return false
		}
		*inListBlock = false
	}

	if !s.yieldEvent(ctx, yield, Event{Kind: StartParagraph}) {
		return false
	}

	if !s.processParagraphContent(ctx, paragraph, book, yield) {
		return false
	}

	return s.yieldEvent(ctx, yield, Event{Kind: EndParagraph})
}

// processParagraphContent handles paragraph text and formatting with chunking
func (s *Streamer) processParagraphContent(ctx context.Context, paragraph *models.Paragraph, book *models.Book, yield func(Event) bool) bool {
	var currentStyle StyleFlags

	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			if !s.processTextRun(ctx, element.TextRun, &currentStyle, yield) {
				return false
			}
		} else if element.InlineObjectElement != nil && book != nil {
			if !s.processInlineImage(ctx, element.InlineObjectElement, book, yield) {
				return false
			}
		}
	}

	// Close any remaining formatting
	return s.closeAllFormatting(ctx, currentStyle, yield)
}

// processTextRun handles text content with chunking for large text
func (s *Streamer) processTextRun(ctx context.Context, textRun *models.TextRun, currentStyle *StyleFlags, yield func(Event) bool) bool {
	newStyle := s.convertTextStyle(textRun.TextStyle)

	// Handle style transitions
	if !s.handleStyleTransition(ctx, *currentStyle, newStyle, textRun.TextStyle.Link, yield) {
		return false
	}
	*currentStyle = newStyle

	// Chunk large text content using bytes.Lines for efficiency
	content := textRun.Content
	if len(content) > s.options.ChunkSize {
		return s.processLargeText(ctx, content, yield)
	}

	// Process regular text
	if content != "" {
		trimmed := s.trimContent(content)
		if !s.options.SkipEmpty || trimmed != "" {
			return s.yieldEvent(ctx, yield, Event{
				Kind:        Text,
				TextContent: trimmed,
			})
		}
	}

	return true
}

// processLargeText chunks large text using Go 1.23+ bytes.Lines
func (s *Streamer) processLargeText(ctx context.Context, content string, yield func(Event) bool) bool {
	// Use bytes.Lines for memory-efficient line processing
	data := []byte(content)
	for line := range bytes.Lines(data) {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		trimmed := s.trimContent(string(line))
		if !s.options.SkipEmpty || trimmed != "" {
			if !s.yieldEvent(ctx, yield, Event{
				Kind:        Text,
				TextContent: trimmed,
			}) {
				return false
			}
		}
	}
	return true
}

// Helper methods

func (s *Streamer) yieldEvent(ctx context.Context, yield func(Event) bool, event Event) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return yield(event)
	}
}

func (s *Streamer) generateUniqueSlug(text string) string {
	base := s.slugify(text)
	if count, exists := s.slugCache[base]; exists {
		s.slugCache[base] = count + 1
		return fmt.Sprintf("%s-%d", base, count+1)
	}
	s.slugCache[base] = 0
	return base
}

func (s *Streamer) slugify(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, " ", "-")
	// Keep Unicode letters, numbers, and hyphens
	var result strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// buildChapterMap creates a lookup map for chapters
func (s *Streamer) buildChapterMap(chapters []models.Chapter) map[string]*models.Chapter {
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

// processTable handles table content
func (s *Streamer) processTable(ctx context.Context, table *models.Table, yield func(Event) bool) bool {
	if len(table.TableRows) == 0 {
		return true
	}

	if !s.yieldEvent(ctx, yield, Event{
		Kind:         StartTable,
		TableColumns: int(table.Columns),
		TableRows:    int(table.Rows),
	}) {
		return false
	}

	for _, row := range table.TableRows {
		if !s.yieldEvent(ctx, yield, Event{Kind: StartTableRow}) {
			return false
		}
		for _, cell := range row.TableCells {
			if !s.yieldEvent(ctx, yield, Event{Kind: StartTableCell}) {
				return false
			}
			for _, element := range cell.Content {
				if element.Paragraph != nil {
					if !s.processParagraphContent(ctx, element.Paragraph, nil, yield) {
						return false
					}
				}
			}
			if !s.yieldEvent(ctx, yield, Event{Kind: EndTableCell}) {
				return false
			}
		}
		if !s.yieldEvent(ctx, yield, Event{Kind: EndTableRow}) {
			return false
		}
	}

	return s.yieldEvent(ctx, yield, Event{Kind: EndTable})
}

// processChapters handles chapter-based content structure
func (s *Streamer) processChapters(ctx context.Context, chapters []models.Chapter, yield func(Event) bool) {
	for _, chapter := range chapters {
		if !s.processChapter(ctx, &chapter, 2, yield) {
			return
		}
	}
}

// processChapter handles individual chapter with proper hierarchical depth
func (s *Streamer) processChapter(ctx context.Context, chapter *models.Chapter, depth int, yield func(Event) bool) bool {
	// Process main chapter at the current depth
	if !s.yieldHeading(ctx, depth, chapter.Title, yield) {
		return false
	}

	// Process subchapters recursively with incremented depth
	for _, subChapter := range chapter.SubChapters {
		if !s.processChapter(ctx, &subChapter, depth+1, yield) {
			return false
		}
	}

	return true
}

// extractParagraphText extracts all text from a paragraph
func (s *Streamer) extractParagraphText(paragraph *models.Paragraph) string {
	var text strings.Builder
	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			text.WriteString(element.TextRun.Content)
		}
	}
	return strings.TrimSpace(text.String())
}

// hasInlineObjects checks if paragraph contains inline objects
func (s *Streamer) hasInlineObjects(paragraph *models.Paragraph) bool {
	for _, element := range paragraph.Elements {
		if element.InlineObjectElement != nil {
			return true
		}
	}
	return false
}

// isHeading checks if paragraph is a heading
func (s *Streamer) isHeading(paragraph *models.Paragraph) bool {
	return strings.HasPrefix(paragraph.ParagraphStyle.NamedStyleType, "HEADING_")
}

// getHeadingLevel converts named style to heading level
func (s *Streamer) getHeadingLevel(namedStyle string) int {
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

// processInlineImage handles inline image elements with meaningful alt text
func (s *Streamer) processInlineImage(ctx context.Context, inlineObj *models.InlineObjectElement, book *models.Book, yield func(Event) bool) bool {
	if book.InlineObjectMap != nil {
		if imageURL, exists := book.InlineObjectMap[inlineObj.InlineObjectID]; exists {
			// Extract meaningful alt text from the document's inline objects
			altText := s.extractImageAltText(inlineObj.InlineObjectID, book)

			return s.yieldEvent(ctx, yield, Event{
				Kind:     Image,
				ImageURL: imageURL,
				ImageAlt: altText,
			})
		}
	}
	return true
}

// extractImageAltText extracts meaningful alt text for an inline image
func (s *Streamer) extractImageAltText(objectID string, book *models.Book) string {
	// First check if we have document content with inline objects
	if book.Content != nil && book.Content.Document != nil {
		if inlineObj, exists := book.Content.Document.InlineObjects[objectID]; exists {
			// Try to get alt text from embedded object title or description
			if inlineObj.InlineObjectProperties.EmbeddedObject.Title != nil &&
				*inlineObj.InlineObjectProperties.EmbeddedObject.Title != "" {
				return *inlineObj.InlineObjectProperties.EmbeddedObject.Title
			}

			if inlineObj.InlineObjectProperties.EmbeddedObject.Description != nil &&
				*inlineObj.InlineObjectProperties.EmbeddedObject.Description != "" {
				return *inlineObj.InlineObjectProperties.EmbeddedObject.Description
			}
		}
	}

	// Fallback: try to create descriptive alt text from object ID or use generic description
	if objectID != "" {
		return "Image: " + objectID
	}

	// Final fallback for accessibility compliance
	return "Embedded image"
}

// closeAllFormatting closes any remaining formatting
func (s *Streamer) closeAllFormatting(ctx context.Context, currentStyle StyleFlags, yield func(Event) bool) bool {
	if currentStyle == 0 {
		return true
	}

	// Close in reverse precedence order
	precedenceOrder := []StyleFlags{Link, Bold, Italic, Underline, Strike, Highlight, Sub, Sup}
	for i := len(precedenceOrder) - 1; i >= 0; i-- {
		style := precedenceOrder[i]
		if currentStyle&style != 0 {
			if !s.yieldEvent(ctx, yield, Event{
				Kind:  EndFormatting,
				Style: style,
			}) {
				return false
			}
		}
	}
	return true
}

// convertTextStyle converts models.TextStyle to StyleFlags
func (s *Streamer) convertTextStyle(textStyle models.TextStyle) StyleFlags {
	var style StyleFlags

	if textStyle.Bold != nil && *textStyle.Bold {
		style |= Bold
	}
	if textStyle.Italic != nil && *textStyle.Italic {
		style |= Italic
	}
	if textStyle.Underline != nil && *textStyle.Underline {
		style |= Underline
	}
	if textStyle.Strikethrough != nil && *textStyle.Strikethrough {
		style |= Strike
	}
	if textStyle.SmallCaps != nil && *textStyle.SmallCaps {
		style |= Highlight // Map small caps to highlight
	}
	if textStyle.Link != nil && textStyle.Link.URL != nil && *textStyle.Link.URL != "" {
		style |= Link
	}

	return style
}

// handleStyleTransition manages style changes
func (s *Streamer) handleStyleTransition(ctx context.Context, currentStyle, newStyle StyleFlags, link *models.Link, yield func(Event) bool) bool {
	if currentStyle == newStyle {
		return true
	}

	changed := currentStyle ^ newStyle
	closing := currentStyle & changed
	opening := newStyle & changed

	var linkURL string
	if link != nil && link.URL != nil {
		linkURL = *link.URL
	}

	// Close styles in reverse order
	precedenceOrder := []StyleFlags{Link, Bold, Italic, Underline, Strike, Highlight, Sub, Sup}
	for i := len(precedenceOrder) - 1; i >= 0; i-- {
		style := precedenceOrder[i]
		if closing&style != 0 {
			if !s.yieldEvent(ctx, yield, Event{
				Kind:    EndFormatting,
				Style:   style,
				LinkURL: linkURL,
			}) {
				return false
			}
		}
	}

	// Open styles in forward order
	for _, style := range precedenceOrder {
		if opening&style != 0 {
			if !s.yieldEvent(ctx, yield, Event{
				Kind:    StartFormatting,
				Style:   style,
				LinkURL: linkURL,
			}) {
				return false
			}
		}
	}

	return true
}

// trimContent performs conservative content trimming
func (s *Streamer) trimContent(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}

	// Replace multiple consecutive newlines with double newlines
	for strings.Contains(trimmed, "\n\n\n") {
		trimmed = strings.ReplaceAll(trimmed, "\n\n\n", "\n\n")
	}

	// Remove trailing newlines
	trimmed = strings.TrimRight(trimmed, "\n\r")
	return trimmed
}
