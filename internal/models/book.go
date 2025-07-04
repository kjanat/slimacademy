// Package models defines the core data structures for SlimAcademy books,
// including books, chapters, documents, paragraphs, tables, and all related
// content types used throughout the transformation pipeline.
package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const customTimeLayout = "2006-01-02 15:04:05"

// normalizeImageURL ensures image URLs have proper schema and host
func normalizeImageURL(url string) string {
	if url == "" {
		return url
	}

	// If URL already has a schema (http:// or https://), return as-is
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}

	// If URL starts with //, add https: prefix
	if strings.HasPrefix(url, "//") {
		return "https:" + url
	}

	// If URL starts with /, prepend the SlimAcademy API host
	if strings.HasPrefix(url, "/") {
		return "https://api.slimacademy.nl" + url
	}

	// For relative URLs without leading slash, prepend full base URL
	return "https://api.slimacademy.nl/" + url
}

// Supplement represents supplementary material associated with a book
type Supplement struct {
	// Most supplements are strings, but can be flexible for future expansion
	Value any `json:"-"`
}

// UnmarshalJSON handles flexible supplement types
func (s *Supplement) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first (most common case)
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.Value = str
		return nil
	}

	// Fallback to interface{} for other types
	return json.Unmarshal(data, &s.Value)
}

// MarshalJSON marshals the supplement value
func (s Supplement) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}

// String returns the string representation of the supplement
func (s Supplement) String() string {
	if str, ok := s.Value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", s.Value)
}

// FormulaImage represents a formula image associated with a book
type FormulaImage struct {
	// Most formula images are objects, but can be flexible for future expansion
	Value any `json:"-"`
}

// UnmarshalJSON handles flexible formula image types
func (f *FormulaImage) UnmarshalJSON(data []byte) error {
	// For now, accept any JSON value since the structure is not well-defined
	return json.Unmarshal(data, &f.Value)
}

// MarshalJSON marshals the formula image value
func (f FormulaImage) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

// CustomTime handles the custom time format used in SlimAcademy JSON data
type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		// Explicitly initialize Time field to zero value when null
		ct.Time = time.Time{}
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
	ID                 int64          `json:"id"`
	Title              string         `json:"title"`
	Description        string         `json:"description"`
	AvailableDate      string         `json:"availableDate"`
	ExamDate           string         `json:"examDate"`
	BachelorYearNumber string         `json:"bachelorYearNumber"`
	CollegeStartYear   int64          `json:"collegeStartYear"`
	ShopURL            string         `json:"shopUrl"`
	IsPurchased        BoolInt        `json:"isPurchased"`
	LastOpenedAt       *CustomTime    `json:"lastOpenedAt"`
	ReadProgress       *int64         `json:"readProgress"`
	PageCount          int64          `json:"pageCount"`
	ReadPageCount      *int64         `json:"readPageCount"`
	ReadPercentage     *float64       `json:"readPercentage"`
	HasFreeChapters    BoolInt        `json:"hasFreeChapters"`
	Supplements        []Supplement   `json:"supplements"`
	Images             []BookImage    `json:"images"`
	FormulasImages     []FormulaImage `json:"formulasImages"`
	Periods            []string       `json:"periods"`

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

// UnmarshalJSON customizes JSON unmarshaling for BookImage to normalize image URLs
func (bi *BookImage) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with the same fields but without custom UnmarshalJSON
	type TempBookImage BookImage
	var temp TempBookImage

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy all fields
	*bi = BookImage(temp)

	// Normalize the image URL
	bi.ImageURL = normalizeImageURL(bi.ImageURL)

	return nil
}

// UnmarshalBook unmarshals JSON data into a Book
func UnmarshalBook(data []byte) (*Book, error) {
	var book Book
	err := json.Unmarshal(data, &book)
	return &book, err
}
