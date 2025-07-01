package transformer

import (
	"strings"

	"github.com/kjanat/slimacademy/internal/models"
)

type Transformer struct{}

func NewTransformer() *Transformer {
	return &Transformer{}
}

func (t *Transformer) Transform(book *models.Book) (*models.Book, error) {
	transformedBook := *book

	t.processContent(&transformedBook)
	t.processImages(&transformedBook)
	t.buildInlineObjectsMapping(&transformedBook)
	t.buildChapterMapping(&transformedBook)

	return &transformedBook, nil
}

func (t *Transformer) processContent(book *models.Book) {
	if book.Content.Body.Content == nil {
		return
	}

	for i := range book.Content.Body.Content {
		element := &book.Content.Body.Content[i]
		if element.Paragraph != nil {
			t.processParagraph(element.Paragraph)
		}
	}
}

func (t *Transformer) processParagraph(paragraph *models.Paragraph) {
	for i := range paragraph.Elements {
		element := &paragraph.Elements[i]
		if element.TextRun != nil {
			element.TextRun.Content = t.cleanText(element.TextRun.Content)
		}
	}
	
	// Consolidate fragmented TextRuns with compatible formatting
	t.consolidateTextRuns(paragraph)
}

func (t *Transformer) cleanText(text string) string {
	// Remove carriage returns which are document artifacts
	text = strings.ReplaceAll(text, "\r", "")
	
	// Only trim leading/trailing whitespace if the text consists entirely of whitespace
	if strings.TrimSpace(text) == "" {
		// Empty or whitespace-only content can be trimmed
		return strings.TrimSpace(text)
	}
	
	// Always trim newlines and tabs - these are clearly document artifacts
	text = strings.Trim(text, "\n\t")
	
	// Check if we have multiple spaces or mixed whitespace (document artifacts)
	// vs single structural spaces that should be preserved
	
	original := text
	fullyTrimmed := strings.TrimSpace(text)
	
	// If trimming removes more than one character from each side,
	// it's likely document formatting, so fully trim
	leftTrimmed := strings.TrimLeft(text, " ")
	rightTrimmed := strings.TrimRight(text, " ")
	
	leftSpaceCount := len(text) - len(leftTrimmed)
	rightSpaceCount := len(text) - len(rightTrimmed)
	
	// If multiple spaces on either side, or any tabs, treat as document artifact
	if leftSpaceCount > 1 || rightSpaceCount > 1 || 
	   strings.Contains(original, "\t") || 
	   strings.Contains(original, "\n") {
		return fullyTrimmed
	}
	
	// Otherwise preserve single leading/trailing spaces as they may be structural
	return text
}

// consolidateTextRuns merges consecutive TextRun elements with compatible formatting
func (t *Transformer) consolidateTextRuns(paragraph *models.Paragraph) {
	if len(paragraph.Elements) <= 1 {
		return
	}

	consolidated := make([]models.Element, 0, len(paragraph.Elements))
	
	for i := 0; i < len(paragraph.Elements); i++ {
		currentElement := paragraph.Elements[i]
		
		// If this is not a TextRun, add it as-is and continue
		if currentElement.TextRun == nil {
			consolidated = append(consolidated, currentElement)
			continue
		}
		
		// Start with the current TextRun
		consolidatedTextRun := *currentElement.TextRun
		
		// Look ahead to find consecutive compatible TextRuns
		j := i + 1
		for j < len(paragraph.Elements) {
			nextElement := paragraph.Elements[j]
			
			// Stop if we hit a non-TextRun element
			if nextElement.TextRun == nil {
				break
			}
			
			// Check if the TextRuns are compatible for merging
			if t.areTextRunsCompatible(currentElement.TextRun, nextElement.TextRun) {
				// Merge the content
				consolidatedTextRun.Content += nextElement.TextRun.Content
				
				// Merge font properties if missing
				t.mergeFontProperties(&consolidatedTextRun.TextStyle, &nextElement.TextRun.TextStyle)
				
				j++
			} else {
				break
			}
		}
		
		// Add the consolidated TextRun
		consolidated = append(consolidated, models.Element{
			TextRun: &consolidatedTextRun,
		})
		
		// Skip the elements we've already processed
		i = j - 1
	}
	
	paragraph.Elements = consolidated
}

// areTextRunsCompatible checks if two TextRuns can be merged based on their formatting
func (t *Transformer) areTextRunsCompatible(textRun1, textRun2 *models.TextRun) bool {
	style1 := textRun1.TextStyle
	style2 := textRun2.TextStyle
	
	// Only consolidate if at least one TextRun has some formatting applied
	// This prevents consolidating plain text that was intentionally separate
	if !t.hasFormatting(&style1) && !t.hasFormatting(&style2) {
		return false
	}
	
	// Check bold formatting compatibility
	if !t.areBoolPointersCompatible(style1.Bold, style2.Bold) {
		return false
	}
	
	// Check italic formatting compatibility
	if !t.areBoolPointersCompatible(style1.Italic, style2.Italic) {
		return false
	}
	
	// Check underline formatting compatibility
	if !t.areBoolPointersCompatible(style1.Underline, style2.Underline) {
		return false
	}
	
	// Check strikethrough formatting compatibility
	if !t.areBoolPointersCompatible(style1.Strikethrough, style2.Strikethrough) {
		return false
	}
	
	// Check smallCaps formatting compatibility
	if !t.areBoolPointersCompatible(style1.SmallCaps, style2.SmallCaps) {
		return false
	}
	
	// Check link compatibility - must be exact match or both nil
	if !t.areLinksCompatible(style1.Link, style2.Link) {
		return false
	}
	
	// Check font size compatibility
	if !t.areFontSizesCompatible(style1.FontSize, style2.FontSize) {
		return false
	}
	
	// Check font family compatibility
	if !t.areFontFamiliesCompatible(style1.WeightedFontFamily, style2.WeightedFontFamily) {
		return false
	}
	
	return true
}

// hasFormatting checks if a TextStyle has any formatting applied
func (t *Transformer) hasFormatting(style *models.TextStyle) bool {
	return (style.Bold != nil && *style.Bold) ||
		(style.Italic != nil && *style.Italic) ||
		(style.Underline != nil && *style.Underline) ||
		(style.Strikethrough != nil && *style.Strikethrough) ||
		(style.SmallCaps != nil && *style.SmallCaps) ||
		(style.Link != nil && style.Link.URL != "") ||
		(style.FontSize != nil) ||
		(style.WeightedFontFamily != nil)
}

// areBoolPointersCompatible checks if two bool pointers are compatible for merging
func (t *Transformer) areBoolPointersCompatible(ptr1, ptr2 *bool) bool {
	// Both nil - compatible
	if ptr1 == nil && ptr2 == nil {
		return true
	}
	
	// One nil, one not nil but false - compatible (nil is treated as false)
	if ptr1 == nil && ptr2 != nil && !*ptr2 {
		return true
	}
	if ptr2 == nil && ptr1 != nil && !*ptr1 {
		return true
	}
	
	// Both non-nil - must have same value
	if ptr1 != nil && ptr2 != nil {
		return *ptr1 == *ptr2
	}
	
	// One nil and one true - not compatible
	return false
}

// areLinksCompatible checks if two links are compatible for merging
func (t *Transformer) areLinksCompatible(link1, link2 *models.Link) bool {
	// Both nil - compatible
	if link1 == nil && link2 == nil {
		return true
	}
	
	// One nil, one not nil - not compatible
	if link1 == nil || link2 == nil {
		return false
	}
	
	// Both non-nil - URLs must match exactly
	return link1.URL == link2.URL
}

// areFontSizesCompatible checks if two font sizes are compatible for merging
func (t *Transformer) areFontSizesCompatible(fontSize1, fontSize2 *models.FontSize) bool {
	// Both nil - compatible
	if fontSize1 == nil && fontSize2 == nil {
		return true
	}
	
	// One nil, one not nil - compatible (nil can inherit from non-nil)
	if fontSize1 == nil || fontSize2 == nil {
		return true
	}
	
	// Both non-nil - magnitude and unit must match exactly
	return fontSize1.Magnitude == fontSize2.Magnitude && fontSize1.Unit == fontSize2.Unit
}

// areFontFamiliesCompatible checks if two font families are compatible for merging
func (t *Transformer) areFontFamiliesCompatible(fontFamily1, fontFamily2 *models.WeightedFontFamily) bool {
	// Both nil - compatible
	if fontFamily1 == nil && fontFamily2 == nil {
		return true
	}
	
	// One nil, one not nil - compatible (nil can inherit from non-nil)
	if fontFamily1 == nil || fontFamily2 == nil {
		return true
	}
	
	// Both non-nil - family and weight must match exactly
	return fontFamily1.FontFamily == fontFamily2.FontFamily && fontFamily1.Weight == fontFamily2.Weight
}

// mergeFontProperties merges font properties from source to target, filling in missing properties
func (t *Transformer) mergeFontProperties(target, source *models.TextStyle) {
	// Merge FontSize if target is missing it
	if target.FontSize == nil && source.FontSize != nil {
		target.FontSize = source.FontSize
	}
	
	// Merge WeightedFontFamily if target is missing it
	if target.WeightedFontFamily == nil && source.WeightedFontFamily != nil {
		target.WeightedFontFamily = source.WeightedFontFamily
	}
}

func (t *Transformer) processImages(book *models.Book) {
	for i := range book.Images {
		image := &book.Images[i]
		image.ImageURL = t.constructImageURL(image.ImageURL)
	}
}

func (t *Transformer) constructImageURL(relativePath string) string {
	const baseURL = "https://api.slimacademy.nl"
	
	// Remove leading slashes and backslashes
	relativePath = strings.TrimLeft(relativePath, "/\\")
	relativePath = strings.ReplaceAll(relativePath, "\\/", "/")
	
	if relativePath == "" {
		return ""
	}
	
	return baseURL + "/" + relativePath
}

func (t *Transformer) buildInlineObjectsMapping(book *models.Book) {
	if book.Content.InlineObjects == nil {
		return
	}
	
	// Create a map to store inline object ID to image URL mapping
	inlineObjectMap := make(map[string]string)
	
	for objectId, objectData := range book.Content.InlineObjects {
		// Parse the inline object data to extract image URL
		if objectDataMap, ok := objectData.(map[string]interface{}); ok {
			if props, ok := objectDataMap["inlineObjectProperties"].(map[string]interface{}); ok {
				if embeddedObj, ok := props["embeddedObject"].(map[string]interface{}); ok {
					if imageProps, ok := embeddedObj["imageProperties"].(map[string]interface{}); ok {
						if contentUri, ok := imageProps["contentUri"].(string); ok {
							inlineObjectMap[objectId] = contentUri
						}
					}
				}
			}
		}
	}
	
	// Store the mapping in the book for use by exporters
	if len(inlineObjectMap) > 0 {
		book.InlineObjectMap = inlineObjectMap
	}
}

func (t *Transformer) buildChapterMapping(book *models.Book) {
	chapterMap := make(map[string]*models.Chapter)

	for i := range book.Chapters {
		chapter := &book.Chapters[i]
		chapterMap[chapter.GDocsChapterID] = chapter

		for j := range chapter.SubChapters {
			subChapter := &chapter.SubChapters[j]
			chapterMap[subChapter.GDocsChapterID] = subChapter
		}
	}

	for i := range book.Content.Body.Content {
		element := &book.Content.Body.Content[i]
		if element.Paragraph != nil && element.Paragraph.ParagraphStyle.HeadingID != nil {
			if chapter, exists := chapterMap[*element.Paragraph.ParagraphStyle.HeadingID]; exists {
				t.linkParagraphToChapter(element.Paragraph, chapter)
			}
		}
	}
}

func (t *Transformer) linkParagraphToChapter(paragraph *models.Paragraph, chapter *models.Chapter) {
}

func (t *Transformer) GetPlainText(book *models.Book) string {
	var text strings.Builder

	for _, element := range book.Content.Body.Content {
		if element.Paragraph != nil {
			for _, elem := range element.Paragraph.Elements {
				if elem.TextRun != nil {
					text.WriteString(elem.TextRun.Content)
				}
			}
		}
	}

	return text.String()
}

func (t *Transformer) GetChapterText(book *models.Book, chapterID string) string {
	var text strings.Builder

	for _, element := range book.Content.Body.Content {
		if element.Paragraph != nil && element.Paragraph.ParagraphStyle.HeadingID != nil {
			if *element.Paragraph.ParagraphStyle.HeadingID == chapterID {
				for _, elem := range element.Paragraph.Elements {
					if elem.TextRun != nil {
						text.WriteString(elem.TextRun.Content)
					}
				}
			}
		}
	}

	return text.String()
}
