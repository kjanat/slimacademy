package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const customTimeLayout = "2006-01-02 15:04:05"

// CustomTime handles the custom time format used in SlimAcademy JSON data
type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		return nil
	}

	// Try SlimAcademy format first
	if t, err := time.Parse(customTimeLayout, s); err == nil {
		ct.Time = t
		return nil
	}

	// Fall back to RFC3339 format for tests
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		ct.Time = t
		return nil
	}

	// Try RFC3339 without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		ct.Time = t
		return nil
	}

	return fmt.Errorf("cannot parse time %q", s)
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + ct.Time.Format(customTimeLayout) + `"`), nil
}

// Book represents a book/document with metadata
type Book struct {
	// Metadata fields from {id}.json
	ID                 int64       `json:"id"`
	Title              string      `json:"title"`
	Description        string      `json:"description"`
	AvailableDate      string      `json:"availableDate"`
	ExamDate           string      `json:"examDate"`
	BachelorYearNumber string      `json:"bachelorYearNumber"`
	CollegeStartYear   int64       `json:"collegeStartYear"`
	ShopURL            string      `json:"shopUrl"`
	IsPurchased        BoolInt     `json:"isPurchased"`
	LastOpenedAt       *time.Time  `json:"lastOpenedAt"`
	ReadProgress       *int64      `json:"readProgress"`
	PageCount          int64       `json:"pageCount"`
	ReadPageCount      any         `json:"readPageCount"`
	ReadPercentage     any         `json:"readPercentage"`
	HasFreeChapters    BoolInt     `json:"hasFreeChapters"`
	Supplements        []any       `json:"supplements"`
	Images             []BookImage `json:"images"`
	FormulasImages     []any       `json:"formulasImages"`
	Periods            []string    `json:"periods"`

	// Additional fields populated from separate JSON files
	Chapters        []Chapter         `json:"-"` // From chapters.json
	Content         *Content          `json:"-"` // From content.json
	InlineObjectMap map[string]string `json:"-"` // Computed map of inline object ID to image URL
}

// BookImage represents an image associated with a book
type BookImage struct {
	ID        int64      `json:"id"`
	SummaryID int64      `json:"summaryId"`
	CreatedAt CustomTime `json:"createdAt"`
	ObjectID  string     `json:"objectId"`
	MIMEType  string     `json:"mimeType"`
	ImageURL  string     `json:"imageUrl"`
}

// UnmarshalBook unmarshals JSON data into a Book
func UnmarshalBook(data []byte) (*Book, error) {
	var book Book
	err := json.Unmarshal(data, &book)
	return &book, err
}
