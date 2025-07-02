package models

import (
	"encoding/json"
	"time"
)

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
	IsPurchased        int64       `json:"isPurchased"`
	LastOpenedAt       *time.Time  `json:"lastOpenedAt"`
	ReadProgress       *int64      `json:"readProgress"`
	PageCount          int64       `json:"pageCount"`
	ReadPageCount      any         `json:"readPageCount"`
	ReadPercentage     any         `json:"readPercentage"`
	HasFreeChapters    int64       `json:"hasFreeChapters"`
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
	ID        int64     `json:"id"`
	SummaryID int64     `json:"summaryId"`
	CreatedAt time.Time `json:"createdAt"`
	ObjectID  string    `json:"objectId"`
	MIMEType  string    `json:"mimeType"`
	ImageURL  string    `json:"imageUrl"`
}

// UnmarshalBook unmarshals JSON data into a Book
func UnmarshalBook(data []byte) (*Book, error) {
	var book Book
	err := json.Unmarshal(data, &book)
	return &book, err
}
