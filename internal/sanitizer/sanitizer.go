package sanitizer

import (
	"fmt"
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
	s.warnings = nil // Reset warnings

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
	if book.Content.Body.Content == nil {
		return
	}

	for i := range book.Content.Body.Content {
		element := &book.Content.Body.Content[i]
		location := fmt.Sprintf("content[%d]", i)

		if element.Paragraph != nil {
			s.sanitizeParagraph(element.Paragraph, location)
		} else if element.Table != nil {
			s.sanitizeTable(element.Table, location)
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
	if link.URL == "" {
		s.addWarning(location, "empty link URL", "", "[EMPTY LINK]")
		return
	}

	// Basic URL validation - check for common patterns
	url := strings.TrimSpace(link.URL)
	if url != link.URL {
		s.addWarning(location, "link URL has whitespace", link.URL, url)
		link.URL = url
	}

	// Check for HTML tags in URL (common issue)
	if strings.Contains(url, "<") || strings.Contains(url, ">") {
		// Remove all HTML-like tags from URL
		cleaned := strings.ReplaceAll(strings.ReplaceAll(url, "<", ""), ">", "")
		if cleaned != url {
			s.addWarning(location, "HTML tags in URL", url, cleaned)
			link.URL = cleaned
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
	if book.ReadPageCount != nil {
		readPageCount := *book.ReadPageCount
		copy.ReadPageCount = &readPageCount
	}
	if book.ReadPercentage != nil {
		readPercentage := *book.ReadPercentage
		copy.ReadPercentage = &readPercentage
	}

	// Copy slices
	copy.Supplements = make([]string, len(book.Supplements))
	copy.Supplements = append(copy.Supplements[:0], book.Supplements...)

	copy.FormulasImages = make([]string, len(book.FormulasImages))
	copy.FormulasImages = append(copy.FormulasImages[:0], book.FormulasImages...)

	copy.Periods = make([]string, len(book.Periods))
	copy.Periods = append(copy.Periods[:0], book.Periods...)

	// Deep copy Images
	copy.Images = make([]models.Image, len(book.Images))
	for i, img := range book.Images {
		copy.Images[i] = models.Image{
			ID:        img.ID,
			SummaryID: img.SummaryID,
			CreatedAt: img.CreatedAt,
			ObjectID:  img.ObjectID,
			MimeType:  img.MimeType,
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
		for k, v := range book.InlineObjectMap {
			copy.InlineObjectMap[k] = v
		}
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
func (s *Sanitizer) deepCopyContent(content models.Content) models.Content {
	copy := models.Content{
		DocumentID:          content.DocumentID,
		RevisionID:          content.RevisionID,
		SuggestionsViewMode: content.SuggestionsViewMode,
		Title:               content.Title,
		Body:                s.deepCopyBody(content.Body),
	}

	// Copy DocumentStyle if present
	if content.DocumentStyle != nil {
		docStyle := *content.DocumentStyle
		copy.DocumentStyle = &docStyle
	}

	// Deep copy maps (these contain any interface{} so we'll copy references)
	if content.Headers != nil {
		copy.Headers = make(map[string]any)
		for k, v := range content.Headers {
			copy.Headers[k] = v
		}
	}
	if content.Footers != nil {
		copy.Footers = make(map[string]any)
		for k, v := range content.Footers {
			copy.Footers[k] = v
		}
	}
	if content.InlineObjects != nil {
		copy.InlineObjects = make(map[string]any)
		for k, v := range content.InlineObjects {
			copy.InlineObjects[k] = v
		}
	}
	if content.Lists != nil {
		copy.Lists = make(map[string]any)
		for k, v := range content.Lists {
			copy.Lists[k] = v
		}
	}
	if content.NamedStyles != nil {
		copy.NamedStyles = make(map[string]any)
		for k, v := range content.NamedStyles {
			copy.NamedStyles[k] = v
		}
	}
	if content.PositionedObjects != nil {
		copy.PositionedObjects = make(map[string]any)
		for k, v := range content.PositionedObjects {
			copy.PositionedObjects[k] = v
		}
	}

	return copy
}

// deepCopyBody creates a deep copy of body
func (s *Sanitizer) deepCopyBody(body models.Body) models.Body {
	copy := models.Body{
		Content: make([]models.ContentElement, len(body.Content)),
	}

	for i, element := range body.Content {
		copy.Content[i] = s.deepCopyContentElement(element)
	}

	return copy
}

// deepCopyContentElement creates a deep copy of content element
func (s *Sanitizer) deepCopyContentElement(element models.ContentElement) models.ContentElement {
	copy := models.ContentElement{
		EndIndex: element.EndIndex,
	}

	if element.StartIndex != nil {
		startIndex := *element.StartIndex
		copy.StartIndex = &startIndex
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
		PositionedObjectIds: make([]string, len(paragraph.PositionedObjectIds)),
		Elements:            make([]models.Element, len(paragraph.Elements)),
		ParagraphStyle:      paragraph.ParagraphStyle,
	}

	// Copy PositionedObjectIds
	copy.PositionedObjectIds = append(copy.PositionedObjectIds[:0], paragraph.PositionedObjectIds...)

	// Deep copy Elements
	for i, element := range paragraph.Elements {
		copy.Elements[i] = s.deepCopyElement(element)
	}

	// Copy Bullet if present
	if paragraph.Bullet != nil {
		bullet := models.Bullet{
			ListID: paragraph.Bullet.ListID,
		}
		// Copy NestingLevel pointer
		if paragraph.Bullet.NestingLevel != nil {
			nestingLevel := *paragraph.Bullet.NestingLevel
			bullet.NestingLevel = &nestingLevel
		}
		// Deep copy TextStyle pointer
		if paragraph.Bullet.TextStyle != nil {
			textStyle := s.deepCopyTextStyle(*paragraph.Bullet.TextStyle)
			bullet.TextStyle = &textStyle
		}
		copy.Bullet = &bullet
	}

	return copy
}

// deepCopyElement creates a deep copy of element
func (s *Sanitizer) deepCopyElement(element models.Element) models.Element {
	copy := models.Element{
		EndIndex:   element.EndIndex,
		StartIndex: element.StartIndex,
	}

	if element.TextRun != nil {
		textRun := s.deepCopyTextRun(*element.TextRun)
		copy.TextRun = &textRun
	}

	if element.InlineObjectElement != nil {
		inlineObj := models.InlineObjectElement{
			InlineObjectID: element.InlineObjectElement.InlineObjectID,
			TextStyle:      s.deepCopyTextStyle(element.InlineObjectElement.TextStyle),
		}
		// Copy slices
		inlineObj.SuggestedDeletionIds = make([]string, len(element.InlineObjectElement.SuggestedDeletionIds))
		inlineObj.SuggestedDeletionIds = append(inlineObj.SuggestedDeletionIds[:0], element.InlineObjectElement.SuggestedDeletionIds...)
		inlineObj.SuggestedInsertionIds = make([]string, len(element.InlineObjectElement.SuggestedInsertionIds))
		inlineObj.SuggestedInsertionIds = append(inlineObj.SuggestedInsertionIds[:0], element.InlineObjectElement.SuggestedInsertionIds...)
		copy.InlineObjectElement = &inlineObj
	}

	if element.PageBreak != nil {
		pageBreak := models.PageBreak{
			TextStyle: s.deepCopyTextStyle(element.PageBreak.TextStyle),
		}
		// Copy slices
		pageBreak.SuggestedDeletionIds = make([]string, len(element.PageBreak.SuggestedDeletionIds))
		pageBreak.SuggestedDeletionIds = append(pageBreak.SuggestedDeletionIds[:0], element.PageBreak.SuggestedDeletionIds...)
		pageBreak.SuggestedInsertionIds = make([]string, len(element.PageBreak.SuggestedInsertionIds))
		pageBreak.SuggestedInsertionIds = append(pageBreak.SuggestedInsertionIds[:0], element.PageBreak.SuggestedInsertionIds...)
		copy.PageBreak = &pageBreak
	}

	if element.ColumnBreak != nil {
		columnBreak := models.ColumnBreak{
			TextStyle: s.deepCopyTextStyle(element.ColumnBreak.TextStyle),
		}
		// Copy slices
		columnBreak.SuggestedDeletionIds = make([]string, len(element.ColumnBreak.SuggestedDeletionIds))
		columnBreak.SuggestedDeletionIds = append(columnBreak.SuggestedDeletionIds[:0], element.ColumnBreak.SuggestedDeletionIds...)
		columnBreak.SuggestedInsertionIds = make([]string, len(element.ColumnBreak.SuggestedInsertionIds))
		columnBreak.SuggestedInsertionIds = append(columnBreak.SuggestedInsertionIds[:0], element.ColumnBreak.SuggestedInsertionIds...)
		copy.ColumnBreak = &columnBreak
	}

	if element.FootnoteReference != nil {
		footnote := models.FootnoteReference{
			FootnoteID:     element.FootnoteReference.FootnoteID,
			FootnoteNumber: element.FootnoteReference.FootnoteNumber,
			TextStyle:      s.deepCopyTextStyle(element.FootnoteReference.TextStyle),
		}
		// Copy slices
		footnote.SuggestedDeletionIds = make([]string, len(element.FootnoteReference.SuggestedDeletionIds))
		footnote.SuggestedDeletionIds = append(footnote.SuggestedDeletionIds[:0], element.FootnoteReference.SuggestedDeletionIds...)
		footnote.SuggestedInsertionIds = make([]string, len(element.FootnoteReference.SuggestedInsertionIds))
		footnote.SuggestedInsertionIds = append(footnote.SuggestedInsertionIds[:0], element.FootnoteReference.SuggestedInsertionIds...)
		copy.FootnoteReference = &footnote
	}

	if element.HorizontalRule != nil {
		rule := models.HorizontalRule{
			TextStyle: s.deepCopyTextStyle(element.HorizontalRule.TextStyle),
		}
		// Copy slices
		rule.SuggestedDeletionIds = make([]string, len(element.HorizontalRule.SuggestedDeletionIds))
		rule.SuggestedDeletionIds = append(rule.SuggestedDeletionIds[:0], element.HorizontalRule.SuggestedDeletionIds...)
		rule.SuggestedInsertionIds = make([]string, len(element.HorizontalRule.SuggestedInsertionIds))
		rule.SuggestedInsertionIds = append(rule.SuggestedInsertionIds[:0], element.HorizontalRule.SuggestedInsertionIds...)
		copy.HorizontalRule = &rule
	}

	if element.Equation != nil {
		equation := *element.Equation
		copy.Equation = &equation
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
		fontSize := models.FontSize{
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

	// Copy slices
	copy.SuggestedDeletionIds = make([]string, len(textRun.SuggestedDeletionIds))
	copy.SuggestedDeletionIds = append(copy.SuggestedDeletionIds[:0], textRun.SuggestedDeletionIds...)

	copy.SuggestedInsertionIds = make([]string, len(textRun.SuggestedInsertionIds))
	copy.SuggestedInsertionIds = append(copy.SuggestedInsertionIds[:0], textRun.SuggestedInsertionIds...)

	return copy
}

// deepCopyTable creates a deep copy of table
func (s *Sanitizer) deepCopyTable(table models.Table) models.Table {
	copy := models.Table{
		Columns: table.Columns,
		Rows:    table.Rows,
	}

	// Copy slices
	copy.SuggestedDeletionIds = make([]string, len(table.SuggestedDeletionIds))
	copy.SuggestedDeletionIds = append(copy.SuggestedDeletionIds[:0], table.SuggestedDeletionIds...)

	copy.SuggestedInsertionIds = make([]string, len(table.SuggestedInsertionIds))
	copy.SuggestedInsertionIds = append(copy.SuggestedInsertionIds[:0], table.SuggestedInsertionIds...)

	// Deep copy TableRows
	copy.TableRows = make([]models.TableRow, len(table.TableRows))
	for i, row := range table.TableRows {
		copy.TableRows[i] = s.deepCopyTableRow(row)
	}

	// Copy TableStyle if present
	if table.TableStyle != nil {
		tableStyle := *table.TableStyle
		copy.TableStyle = &tableStyle
	}

	return copy
}

// deepCopyTableRow creates a deep copy of table row
func (s *Sanitizer) deepCopyTableRow(row models.TableRow) models.TableRow {
	copy := models.TableRow{
		EndIndex:   row.EndIndex,
		StartIndex: row.StartIndex,
	}

	// Copy slices
	copy.SuggestedDeletionIds = make([]string, len(row.SuggestedDeletionIds))
	copy.SuggestedDeletionIds = append(copy.SuggestedDeletionIds[:0], row.SuggestedDeletionIds...)

	copy.SuggestedInsertionIds = make([]string, len(row.SuggestedInsertionIds))
	copy.SuggestedInsertionIds = append(copy.SuggestedInsertionIds[:0], row.SuggestedInsertionIds...)

	// Deep copy TableCells
	copy.TableCells = make([]models.TableCell, len(row.TableCells))
	for i, cell := range row.TableCells {
		copy.TableCells[i] = s.deepCopyTableCell(cell)
	}

	// Copy TableRowStyle if present
	if row.TableRowStyle != nil {
		rowStyle := *row.TableRowStyle
		copy.TableRowStyle = &rowStyle
	}

	return copy
}

// deepCopyTableCell creates a deep copy of table cell
func (s *Sanitizer) deepCopyTableCell(cell models.TableCell) models.TableCell {
	copy := models.TableCell{
		EndIndex:   cell.EndIndex,
		StartIndex: cell.StartIndex,
	}

	// Copy slices
	copy.SuggestedDeletionIds = make([]string, len(cell.SuggestedDeletionIds))
	copy.SuggestedDeletionIds = append(copy.SuggestedDeletionIds[:0], cell.SuggestedDeletionIds...)

	copy.SuggestedInsertionIds = make([]string, len(cell.SuggestedInsertionIds))
	copy.SuggestedInsertionIds = append(copy.SuggestedInsertionIds[:0], cell.SuggestedInsertionIds...)

	// Deep copy Content
	copy.Content = make([]models.ContentElement, len(cell.Content))
	for i, element := range cell.Content {
		copy.Content[i] = s.deepCopyContentElement(element)
	}

	// Copy TableCellStyle if present
	if cell.TableCellStyle != nil {
		cellStyle := *cell.TableCellStyle
		copy.TableCellStyle = &cellStyle
	}

	return copy
}
