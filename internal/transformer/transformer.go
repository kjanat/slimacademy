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
}

func (t *Transformer) cleanText(text string) string {
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.TrimSpace(text)
	return text
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