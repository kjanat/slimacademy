package models

import "encoding/json"

// Chapter represents a chapter in a book
type Chapter struct {
	ID              int64     `json:"id"`
	SummaryID       int64     `json:"summaryId"`
	Title           string    `json:"title"`
	IsFree          int64     `json:"isFree"`
	IsSupplement    int64     `json:"isSupplement"`
	IsLocked        int64     `json:"isLocked"`
	IsVisible       int64     `json:"isVisible"`
	ParentChapterID *int64    `json:"parentChapterId"`
	GDocsChapterID  string    `json:"gDocsChapterId"`
	SortIndex       int64     `json:"sortIndex"`
	SubChapters     []Chapter `json:"subChapters"`
}

// UnmarshalChapters unmarshals JSON data into chapters
func UnmarshalChapters(data []byte) ([]Chapter, error) {
	var chapters []Chapter
	err := json.Unmarshal(data, &chapters)
	return chapters, err
}
