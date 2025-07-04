package models

import (
	"encoding/json"
)

// BoolInt represents a boolean value that comes from JSON as 0/1 integer
type BoolInt bool

// UnmarshalJSON converts 0/1 integers to boolean values
func (b *BoolInt) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case float64:
		*b = BoolInt(v != 0)
	case int:
		*b = BoolInt(v != 0)
	case int64:
		*b = BoolInt(v != 0)
	case string:
		// Handle string representation
		if v == "1" || v == "true" {
			*b = BoolInt(true)
		} else {
			*b = BoolInt(false)
		}
	case bool:
		*b = BoolInt(v)
	default:
		*b = BoolInt(false)
	}

	return nil
}

// MarshalJSON converts boolean back to 0/1 for JSON output
func (b BoolInt) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

// Bool returns the boolean value
func (b BoolInt) Bool() bool {
	return bool(b)
}

// String returns string representation
func (b BoolInt) String() string {
	if b {
		return "1"
	}
	return "0"
}

// Int64 returns the int64 representation (0 or 1)
func (b BoolInt) Int64() int64 {
	if b {
		return 1
	}
	return 0
}

// Chapter represents a chapter in a book
type Chapter struct {
	ID              int64     `json:"id"`
	SummaryID       int64     `json:"summaryId"`
	Title           string    `json:"title"`
	IsFree          BoolInt   `json:"isFree"`
	IsSupplement    BoolInt   `json:"isSupplement"`
	IsLocked        BoolInt   `json:"isLocked"`
	IsVisible       BoolInt   `json:"isVisible"`
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
