package models

// InlineObject represents an inline object (like an image)
type InlineObject struct {
	ObjectID               string                 `json:"objectId"`
	InlineObjectProperties InlineObjectProperties `json:"inlineObjectProperties"`
	SuggestedDeletionIDs   any                    `json:"suggestedDeletionIds"`
	SuggestedInsertionID   any                    `json:"suggestedInsertionId"`
}

// InlineObjectProperties represents properties of an inline object
type InlineObjectProperties struct {
	EmbeddedObject EmbeddedObject `json:"embeddedObject"`
}

// PositionedObject represents a positioned object
type PositionedObject struct {
	ObjectID                   string                     `json:"objectId"`
	PositionedObjectProperties PositionedObjectProperties `json:"positionedObjectProperties"`
	SuggestedDeletionIDs       any                        `json:"suggestedDeletionIds"`
	SuggestedInsertionID       any                        `json:"suggestedInsertionId"`
}

// PositionedObjectProperties represents properties of a positioned object
type PositionedObjectProperties struct {
	Positioning    Positioning    `json:"positioning"`
	EmbeddedObject EmbeddedObject `json:"embeddedObject"`
}

// Positioning represents positioning information
type Positioning struct {
	Layout     string    `json:"layout"`
	LeftOffset Dimension `json:"leftOffset"`
	TopOffset  Dimension `json:"topOffset"`
}

// EmbeddedObject represents an embedded object
type EmbeddedObject struct {
	Title                *string              `json:"title,omitempty"`
	Description          *string              `json:"description,omitempty"`
	EmbeddedObjectBorder EmbeddedObjectBorder `json:"embeddedObjectBorder"`
	Size                 Size                 `json:"size"`
	MarginTop            Dimension            `json:"marginTop"`
	MarginBottom         Dimension            `json:"marginBottom"`
	MarginRight          Dimension            `json:"marginRight"`
	MarginLeft           Dimension            `json:"marginLeft"`
	ImageProperties      *ImageProperties     `json:"imageProperties,omitempty"`
}

// EmbeddedObjectBorder represents border of an embedded object
type EmbeddedObjectBorder struct {
	Color         *Color    `json:"color,omitempty"`
	Width         Dimension `json:"width"`
	DashStyle     string    `json:"dashStyle"`
	PropertyState string    `json:"propertyState"`
}

// ImageProperties represents properties of an image
type ImageProperties struct {
	ContentURI     string         `json:"contentUri"`
	SourceURI      *string        `json:"sourceUri,omitempty"`
	Brightness     *float64       `json:"brightness,omitempty"`
	Contrast       *float64       `json:"contrast,omitempty"`
	Transparency   *float64       `json:"transparency,omitempty"`
	Angle          *float64       `json:"angle,omitempty"`
	CropProperties CropProperties `json:"cropProperties"`
}

// CropProperties represents image cropping
type CropProperties struct {
	OffsetLeft   *float64 `json:"offsetLeft,omitempty"`
	OffsetRight  *float64 `json:"offsetRight,omitempty"`
	OffsetTop    *float64 `json:"offsetTop,omitempty"`
	OffsetBottom *float64 `json:"offsetBottom,omitempty"`
	Angle        *float64 `json:"angle,omitempty"`
}
