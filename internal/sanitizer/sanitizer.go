// Package sanitizer provides text sanitization and content cleaning for SlimAcademy documents.
// It handles UTF-8 validation, whitespace normalization, and malformed content detection
// with detailed diagnostic reporting.
package sanitizer

import (
	"fmt"
	"maps"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/kjanat/slimacademy/internal/models"
)

// Warning represents a sanitization warning with location info
type Warning struct {
	Location string
	Issue    string
	Original string
	Fixed    string
}

// Result contains the sanitized book and any warnings
type Result struct {
	Book     *models.Book
	Warnings []Warning
}

// Sanitizer cleans and validates document content before event generation
type Sanitizer struct {
	warnings []Warning
}

// NewSanitizer returns a new Sanitizer instance with an initialized empty warnings list.
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		warnings: make([]Warning, 0),
	}
}

// Sanitize processes a book and returns cleaned content with warnings
func (s *Sanitizer) Sanitize(book *models.Book) *Result {
	s.warnings = s.warnings[:0] // Reset warnings slice but keep capacity

	// Handle nil book
	if book == nil {
		return &Result{
			Book:     nil,
			Warnings: s.warnings,
		}
	}

	// Create a copy to avoid mutating the original
	sanitized := s.deepCopyBook(book)

	// Sanitize document content
	s.sanitizeContent(sanitized)
	s.sanitizeChapters(sanitized)
	s.sanitizeImages(sanitized)

	return &Result{
		Book:     sanitized,
		Warnings: s.warnings,
	}
}

// sanitizeContent cleans the document body content
func (s *Sanitizer) sanitizeContent(book *models.Book) {
	if book.Content == nil {
		return
	}

	// Handle document-based content
	if book.Content.Document != nil && book.Content.Document.Body.Content != nil {
		for i := range book.Content.Document.Body.Content {
			element := &book.Content.Document.Body.Content[i]
			location := fmt.Sprintf("content[%d]", i)

			if element.Paragraph != nil {
				s.sanitizeParagraph(element.Paragraph, location)
			} else if element.Table != nil {
				s.sanitizeTable(element.Table, location)
			}
		}
	}

	// Handle chapter-based content
	if book.Content.Chapters != nil {
		for i := range book.Content.Chapters {
			chapter := &book.Content.Chapters[i]
			location := fmt.Sprintf("content.chapters[%d]", i)

			if strings.TrimSpace(chapter.Title) == "" {
				s.addWarning(location, "empty chapter title", chapter.Title, "[EMPTY CHAPTER]")
			}
		}
	}
}

// sanitizeParagraph cleans paragraph content
func (s *Sanitizer) sanitizeParagraph(paragraph *models.Paragraph, location string) {
	for i := range paragraph.Elements {
		element := &paragraph.Elements[i]
		elemLocation := fmt.Sprintf("%s.elements[%d]", location, i)

		if element.TextRun != nil {
			s.sanitizeTextRun(element.TextRun, elemLocation)
		}
	}

	// Check for empty heading payloads
	if s.isHeading(paragraph) {
		text := s.extractText(paragraph)
		if strings.TrimSpace(text) == "" {
			s.addWarning(location, "empty heading payload", text, "[EMPTY HEADING REMOVED]")
		}
	}
}

// sanitizeTextRun cleans text run content
func (s *Sanitizer) sanitizeTextRun(textRun *models.TextRun, location string) {
	original := textRun.Content
	cleaned := s.sanitizeText(original)

	if cleaned != original {
		s.addWarning(location, "text content sanitized", original, cleaned)
		textRun.Content = cleaned
	}

	// Validate link URLs if present
	if textRun.TextStyle.Link != nil {
		s.validateLinkURL(textRun.TextStyle.Link, location)
	}
}

// sanitizeTable cleans table content
func (s *Sanitizer) sanitizeTable(table *models.Table, location string) {
	for rowIdx, row := range table.TableRows {
		for cellIdx, cell := range row.TableCells {
			cellLocation := fmt.Sprintf("%s.row[%d].cell[%d]", location, rowIdx, cellIdx)
			for contentIdx := range cell.Content {
				contentLocation := fmt.Sprintf("%s.content[%d]", cellLocation, contentIdx)
				if cell.Content[contentIdx].Paragraph != nil {
					s.sanitizeParagraph(cell.Content[contentIdx].Paragraph, contentLocation)
				}
			}
		}
	}
}

// sanitizeChapters validates chapter structure
func (s *Sanitizer) sanitizeChapters(book *models.Book) {
	for i := range book.Chapters {
		chapter := &book.Chapters[i]
		location := fmt.Sprintf("chapters[%d]", i)

		if strings.TrimSpace(chapter.Title) == "" {
			s.addWarning(location, "empty chapter title", chapter.Title, "[EMPTY CHAPTER]")
		}

		for j := range chapter.SubChapters {
			subChapter := &chapter.SubChapters[j]
			subLocation := fmt.Sprintf("%s.subchapters[%d]", location, j)

			if strings.TrimSpace(subChapter.Title) == "" {
				s.addWarning(subLocation, "empty subchapter title", subChapter.Title, "[EMPTY SUBCHAPTER]")
			}
		}
	}
}

// sanitizeImages validates image references
func (s *Sanitizer) sanitizeImages(book *models.Book) {
	for i := range book.Images {
		image := &book.Images[i]
		location := fmt.Sprintf("images[%d]", i)

		if image.ImageURL == "" {
			s.addWarning(location, "empty image URL", "", "[MISSING IMAGE]")
		}
	}
}

// sanitizeText removes control characters and normalizes whitespace
func (s *Sanitizer) sanitizeText(text string) string {
	if text == "" {
		return text
	}

	// Remove non-printable control characters except common whitespace
	var cleaned strings.Builder
	cleaned.Grow(len(text))

	for _, r := range text {
		switch {
		case r == '\n' || r == '\t' || r == '\r' || r == ' ':
			// Keep common whitespace
			cleaned.WriteRune(r)
		case unicode.IsControl(r):
			// Remove other control characters
			continue
		case !utf8.ValidRune(r):
			// Remove invalid UTF-8 runes
			continue
		default:
			cleaned.WriteRune(r)
		}
	}

	result := cleaned.String()

	// Normalize excessive whitespace
	result = s.normalizeWhitespace(result)

	return result
}

// normalizeWhitespace removes excessive whitespace while preserving structure
func (s *Sanitizer) normalizeWhitespace(text string) string {
	// Remove carriage returns
	text = strings.ReplaceAll(text, "\r", "")

	// Split into lines to preserve newline structure
	lines := strings.Split(text, "\n")

	// Normalize spaces within each line individually
	for i, line := range lines {
		// Replace multiple spaces with single space within the line
		words := strings.Fields(line)
		if len(words) == 0 {
			lines[i] = ""
		} else {
			lines[i] = strings.Join(words, " ")
		}
	}

	// Reconstruct with newlines preserved
	return strings.Join(lines, "\n")
}

// validateLinkURL checks if a link URL is well-formed
func (s *Sanitizer) validateLinkURL(link *models.Link, location string) {
	if link.URL == nil || *link.URL == "" {
		s.addWarning(location, "empty link URL", "", "[EMPTY LINK]")
		return
	}

	// Basic URL validation - check for common patterns
	url := strings.TrimSpace(*link.URL)
	if url != *link.URL {
		s.addWarning(location, "link URL has whitespace", *link.URL, url)
		link.URL = &url
	}

	// Check for HTML tags in URL (common issue)
	if strings.Contains(url, "<") || strings.Contains(url, ">") {
		// Remove all HTML-like tags from URL
		cleaned := strings.ReplaceAll(strings.ReplaceAll(url, "<", ""), ">", "")
		if cleaned != url {
			s.addWarning(location, "HTML tags in URL", url, cleaned)
			link.URL = &cleaned
		}
	}
}

// Helper methods

func (s *Sanitizer) isHeading(paragraph *models.Paragraph) bool {
	return strings.HasPrefix(paragraph.ParagraphStyle.NamedStyleType, "HEADING_")
}

func (s *Sanitizer) extractText(paragraph *models.Paragraph) string {
	var text strings.Builder
	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			text.WriteString(element.TextRun.Content)
		}
	}
	return text.String()
}

func (s *Sanitizer) addWarning(location, issue, original, fixed string) {
	s.warnings = append(s.warnings, Warning{
		Location: location,
		Issue:    issue,
		Original: original,
		Fixed:    fixed,
	})
}

// deepCopyBook creates a deep copy of the book to avoid mutations
func (s *Sanitizer) deepCopyBook(book *models.Book) *models.Book {
	if book == nil {
		return nil
	}

	// Create new book with copied scalar fields
	copy := &models.Book{
		ID:                 book.ID,
		Title:              book.Title,
		Description:        book.Description,
		AvailableDate:      book.AvailableDate,
		ExamDate:           book.ExamDate,
		BachelorYearNumber: book.BachelorYearNumber,
		CollegeStartYear:   book.CollegeStartYear,
		ShopURL:            book.ShopURL,
		IsPurchased:        book.IsPurchased,
		PageCount:          book.PageCount,
		HasFreeChapters:    book.HasFreeChapters,
	}

	// Copy pointer fields
	if book.LastOpenedAt != nil {
		lastOpened := *book.LastOpenedAt
		copy.LastOpenedAt = &lastOpened
	}
	if book.ReadProgress != nil {
		readProgress := *book.ReadProgress
		copy.ReadProgress = &readProgress
	}
	copy.ReadPageCount = book.ReadPageCount
	copy.ReadPercentage = book.ReadPercentage

	// Copy slices
	copy.Supplements = make([]models.Supplement, len(book.Supplements))
	copy.Supplements = append(copy.Supplements[:0], book.Supplements...)

	copy.FormulasImages = make([]models.FormulaImage, len(book.FormulasImages))
	copy.FormulasImages = append(copy.FormulasImages[:0], book.FormulasImages...)

	copy.Periods = make([]string, len(book.Periods))
	copy.Periods = append(copy.Periods[:0], book.Periods...)

	// Deep copy Images
	copy.Images = make([]models.BookImage, len(book.Images))
	for i, img := range book.Images {
		copy.Images[i] = models.BookImage{
			ID:        img.ID,
			SummaryID: img.SummaryID,
			CreatedAt: img.CreatedAt,
			ObjectID:  img.ObjectID,
			MIMEType:  img.MIMEType,
			ImageURL:  img.ImageURL,
		}
	}

	// Deep copy Chapters
	copy.Chapters = s.deepCopyChapters(book.Chapters)

	// Deep copy Content
	copy.Content = s.deepCopyContent(book.Content)

	// Copy InlineObjectMap
	if book.InlineObjectMap != nil {
		copy.InlineObjectMap = make(map[string]string)
		maps.Copy(copy.InlineObjectMap, book.InlineObjectMap)
	}

	return copy
}

// deepCopyChapters creates a deep copy of chapters slice
func (s *Sanitizer) deepCopyChapters(chapters []models.Chapter) []models.Chapter {
	if chapters == nil {
		return nil
	}

	copy := make([]models.Chapter, len(chapters))
	for i, chapter := range chapters {
		copy[i] = models.Chapter{
			ID:             chapter.ID,
			SummaryID:      chapter.SummaryID,
			Title:          chapter.Title,
			IsFree:         chapter.IsFree,
			IsSupplement:   chapter.IsSupplement,
			IsLocked:       chapter.IsLocked,
			IsVisible:      chapter.IsVisible,
			GDocsChapterID: chapter.GDocsChapterID,
			SortIndex:      chapter.SortIndex,
		}

		// Copy ParentChapterID pointer
		if chapter.ParentChapterID != nil {
			parentID := *chapter.ParentChapterID
			copy[i].ParentChapterID = &parentID
		}

		// Recursively copy SubChapters
		copy[i].SubChapters = s.deepCopyChapters(chapter.SubChapters)
	}

	return copy
}

// deepCopyContent creates a deep copy of content
func (s *Sanitizer) deepCopyContent(content *models.Content) *models.Content {
	if content == nil {
		return nil
	}

	copy := &models.Content{}

	// Copy Document if present
	if content.Document != nil {
		copy.Document = s.deepCopyDocument(content.Document)
	}

	// Copy Chapters if present
	if content.Chapters != nil {
		copy.Chapters = s.deepCopyChapters(content.Chapters)
	}

	return copy
}

// deepCopyDocument creates a deep copy of document
func (s *Sanitizer) deepCopyDocument(doc *models.Document) *models.Document {
	if doc == nil {
		return nil
	}

	copy := &models.Document{
		DocumentID:          doc.DocumentID,
		RevisionID:          doc.RevisionID,
		SuggestionsViewMode: doc.SuggestionsViewMode,
		Title:               doc.Title,
		Body:                s.deepCopyBody(doc.Body),
		DocumentStyle:       doc.DocumentStyle,
	}

	// Deep copy maps (these contain any interface{} so we'll copy references)
	if doc.Headers != nil {
		copy.Headers = make(map[string]models.HeaderFooter)
		maps.Copy(copy.Headers, doc.Headers)
	}
	if doc.Footers != nil {
		copy.Footers = make(map[string]models.HeaderFooter)
		maps.Copy(copy.Footers, doc.Footers)
	}
	if doc.InlineObjects != nil {
		copy.InlineObjects = make(map[string]models.InlineObject)
		maps.Copy(copy.InlineObjects, doc.InlineObjects)
	}
	if doc.Lists != nil {
		copy.Lists = make(map[string]models.List)
		maps.Copy(copy.Lists, doc.Lists)
	}
	if doc.PositionedObjects != nil {
		copy.PositionedObjects = make(map[string]models.PositionedObject)
		maps.Copy(copy.PositionedObjects, doc.PositionedObjects)
	}

	return copy
}

// deepCopyBody creates a deep copy of body
func (s *Sanitizer) deepCopyBody(body models.Body) models.Body {
	copy := models.Body{
		Content: make([]models.StructuralElement, len(body.Content)),
	}

	for i, element := range body.Content {
		copy.Content[i] = s.deepCopyStructuralElement(element)
	}

	return copy
}

// deepCopyStructuralElement creates a deep copy of structural element
func (s *Sanitizer) deepCopyStructuralElement(element models.StructuralElement) models.StructuralElement {
	copy := models.StructuralElement{
		StartIndex: element.StartIndex,
		EndIndex:   element.EndIndex,
	}

	if element.SectionBreak != nil {
		sectionBreak := *element.SectionBreak
		copy.SectionBreak = &sectionBreak
	}

	if element.Paragraph != nil {
		paragraph := s.deepCopyParagraph(*element.Paragraph)
		copy.Paragraph = &paragraph
	}

	if element.Table != nil {
		table := s.deepCopyTable(*element.Table)
		copy.Table = &table
	}

	if element.TableOfContents != nil {
		toc := *element.TableOfContents
		copy.TableOfContents = &toc
	}

	return copy
}

// deepCopyParagraph creates a deep copy of paragraph
func (s *Sanitizer) deepCopyParagraph(paragraph models.Paragraph) models.Paragraph {
	copy := models.Paragraph{
		Elements:            make([]models.ParagraphElement, len(paragraph.Elements)),
		ParagraphStyle:      paragraph.ParagraphStyle,
		PositionedObjectIDs: paragraph.PositionedObjectIDs,
	}

	// Copy Bullet if present
	if paragraph.Bullet != nil {
		bullet := *paragraph.Bullet
		copy.Bullet = &bullet
	}

	// Deep copy Elements
	for i, element := range paragraph.Elements {
		copy.Elements[i] = s.deepCopyParagraphElement(element)
	}

	return copy
}

// deepCopyParagraphElement creates a deep copy of paragraph element
func (s *Sanitizer) deepCopyParagraphElement(element models.ParagraphElement) models.ParagraphElement {
	copy := models.ParagraphElement{
		StartIndex: element.StartIndex,
		EndIndex:   element.EndIndex,
	}

	if element.TextRun != nil {
		textRun := s.deepCopyTextRun(*element.TextRun)
		copy.TextRun = &textRun
	}

	if element.InlineObjectElement != nil {
		inlineObj := models.InlineObjectElement{
			InlineObjectID:        element.InlineObjectElement.InlineObjectID,
			TextStyle:             element.InlineObjectElement.TextStyle,
			SuggestedDeletionIDs:  element.InlineObjectElement.SuggestedDeletionIDs,
			SuggestedInsertionIDs: element.InlineObjectElement.SuggestedInsertionIDs,
		}
		copy.InlineObjectElement = &inlineObj
	}

	if element.PageBreak != nil {
		pageBreak := models.PageBreak{
			TextStyle:             element.PageBreak.TextStyle,
			SuggestedDeletionIDs:  element.PageBreak.SuggestedDeletionIDs,
			SuggestedInsertionIDs: element.PageBreak.SuggestedInsertionIDs,
		}
		copy.PageBreak = &pageBreak
	}

	if element.HorizontalRule != nil {
		rule := models.HorizontalRule{
			TextStyle:             element.HorizontalRule.TextStyle,
			SuggestedDeletionIDs:  element.HorizontalRule.SuggestedDeletionIDs,
			SuggestedInsertionIDs: element.HorizontalRule.SuggestedInsertionIDs,
		}
		copy.HorizontalRule = &rule
	}

	if element.AutoText != nil {
		autoText := models.AutoText{
			Type:                  element.AutoText.Type,
			TextStyle:             element.AutoText.TextStyle,
			SuggestedDeletionIDs:  element.AutoText.SuggestedDeletionIDs,
			SuggestedInsertionIDs: element.AutoText.SuggestedInsertionIDs,
		}
		copy.AutoText = &autoText
	}

	return copy
}

// deepCopyTextStyle creates a deep copy of text style
func (s *Sanitizer) deepCopyTextStyle(textStyle models.TextStyle) models.TextStyle {
	copy := models.TextStyle{
		ForegroundColor: textStyle.ForegroundColor,
		BackgroundColor: textStyle.BackgroundColor,
	}

	// Copy pointer fields
	if textStyle.BaselineOffset != nil {
		baselineOffset := *textStyle.BaselineOffset
		copy.BaselineOffset = &baselineOffset
	}
	if textStyle.Bold != nil {
		bold := *textStyle.Bold
		copy.Bold = &bold
	}
	if textStyle.Italic != nil {
		italic := *textStyle.Italic
		copy.Italic = &italic
	}
	if textStyle.SmallCaps != nil {
		smallCaps := *textStyle.SmallCaps
		copy.SmallCaps = &smallCaps
	}
	if textStyle.Strikethrough != nil {
		strikethrough := *textStyle.Strikethrough
		copy.Strikethrough = &strikethrough
	}
	if textStyle.Underline != nil {
		underline := *textStyle.Underline
		copy.Underline = &underline
	}

	// Copy FontSize if present
	if textStyle.FontSize != nil {
		fontSize := models.Dimension{
			Magnitude: textStyle.FontSize.Magnitude,
			Unit:      textStyle.FontSize.Unit,
		}
		copy.FontSize = &fontSize
	}

	// Copy WeightedFontFamily if present
	if textStyle.WeightedFontFamily != nil {
		fontFamily := models.WeightedFontFamily{
			FontFamily: textStyle.WeightedFontFamily.FontFamily,
			Weight:     textStyle.WeightedFontFamily.Weight,
		}
		copy.WeightedFontFamily = &fontFamily
	}

	// Deep copy Link if present
	if textStyle.Link != nil {
		link := models.Link{
			URL: textStyle.Link.URL,
		}
		if textStyle.Link.BookmarkID != nil {
			bookmarkID := *textStyle.Link.BookmarkID
			link.BookmarkID = &bookmarkID
		}
		if textStyle.Link.HeadingID != nil {
			headingID := *textStyle.Link.HeadingID
			link.HeadingID = &headingID
		}
		copy.Link = &link
	}

	return copy
}

// deepCopyTextRun creates a deep copy of text run
func (s *Sanitizer) deepCopyTextRun(textRun models.TextRun) models.TextRun {
	copy := models.TextRun{
		Content:   textRun.Content,
		TextStyle: s.deepCopyTextStyle(textRun.TextStyle),
	}

	// Copy interface{} fields directly
	copy.SuggestedDeletionIDs = textRun.SuggestedDeletionIDs
	copy.SuggestedInsertionIDs = textRun.SuggestedInsertionIDs

	return copy
}

// deepCopyTable creates a deep copy of table
func (s *Sanitizer) deepCopyTable(table models.Table) models.Table {
	copy := models.Table{
		Columns:               table.Columns,
		Rows:                  table.Rows,
		TableStyle:            table.TableStyle,
		SuggestedDeletionIDs:  table.SuggestedDeletionIDs,
		SuggestedInsertionIDs: table.SuggestedInsertionIDs,
	}

	// Deep copy TableRows
	copy.TableRows = make([]models.TableRow, len(table.TableRows))
	for i, row := range table.TableRows {
		copy.TableRows[i] = s.deepCopyTableRow(row)
	}

	return copy
}

// deepCopyTableRow creates a deep copy of table row
func (s *Sanitizer) deepCopyTableRow(row models.TableRow) models.TableRow {
	copy := models.TableRow{
		StartIndex:            row.StartIndex,
		EndIndex:              row.EndIndex,
		TableRowStyle:         row.TableRowStyle,
		SuggestedDeletionIDs:  row.SuggestedDeletionIDs,
		SuggestedInsertionIDs: row.SuggestedInsertionIDs,
	}

	// Deep copy TableCells
	copy.TableCells = make([]models.TableCell, len(row.TableCells))
	for i, cell := range row.TableCells {
		copy.TableCells[i] = s.deepCopyTableCell(cell)
	}

	return copy
}

// deepCopyTableCell creates a deep copy of table cell
func (s *Sanitizer) deepCopyTableCell(cell models.TableCell) models.TableCell {
	copy := models.TableCell{
		StartIndex:            cell.StartIndex,
		EndIndex:              cell.EndIndex,
		TableCellStyle:        cell.TableCellStyle,
		SuggestedDeletionIDs:  cell.SuggestedDeletionIDs,
		SuggestedInsertionIDs: cell.SuggestedInsertionIDs,
	}

	// Deep copy Content
	copy.Content = make([]models.StructuralElement, len(cell.Content))
	for i, element := range cell.Content {
		copy.Content[i] = s.deepCopyStructuralElement(element)
	}

	return copy
}
