// Package hast provides types and utilities for working with Hypertext Abstract Syntax Trees (HAST).
// HAST is a specification for representing HTML as an abstract syntax tree.
// This package is ported from the slim project with enhancements for SlimAcademy.
//
// HAST specification: https://github.com/syntax-tree/hast
package hast

import (
	"encoding/json"
	"fmt"
)

// Node represents any HAST node.
// This is a Go equivalent of the union type used in TypeScript HAST implementations.
type Node interface {
	// Type returns the type of the HAST node
	Type() string

	// Accept allows visitors to traverse the HAST tree
	Accept(visitor Visitor) error
}

// Visitor defines the interface for traversing HAST trees
type Visitor interface {
	VisitRoot(*Root) error
	VisitElement(*Element) error
	VisitText(*Text) error
	VisitComment(*Comment) error
}

// Root represents the root of a HAST tree.
// https://github.com/syntax-tree/hast#root
type Root struct {
	NodeType string `json:"type"`
	Children []Node `json:"children,omitempty"`
}

// Type returns the HAST node type
func (r *Root) Type() string {
	return "root"
}

// Accept implements the Visitor pattern
func (r *Root) Accept(visitor Visitor) error {
	return visitor.VisitRoot(r)
}

// Element represents a HAST element (HTML tag).
// https://github.com/syntax-tree/hast#element
type Element struct {
	NodeType   string         `json:"type"`
	TagName    string         `json:"tagName"`
	Properties map[string]any `json:"properties,omitempty"`
	Children   []Node         `json:"children,omitempty"`
}

// Type returns the HAST node type
func (e *Element) Type() string {
	return "element"
}

// Accept implements the Visitor pattern
func (e *Element) Accept(visitor Visitor) error {
	return visitor.VisitElement(e)
}

// Text represents a HAST text node.
// https://github.com/syntax-tree/hast#text
type Text struct {
	NodeType string `json:"type"`
	Value    string `json:"value"`
}

// Type returns the HAST node type
func (t *Text) Type() string {
	return "text"
}

// Accept implements the Visitor pattern
func (t *Text) Accept(visitor Visitor) error {
	return visitor.VisitText(t)
}

// Comment represents a HAST comment node.
// https://github.com/syntax-tree/hast#comment
type Comment struct {
	NodeType string `json:"type"`
	Value    string `json:"value"`
}

// Type returns the HAST node type
func (c *Comment) Type() string {
	return "comment"
}

// Accept implements the Visitor pattern
func (c *Comment) Accept(visitor Visitor) error {
	return visitor.VisitComment(c)
}

// NewRoot creates a new HAST root node
func NewRoot() *Root {
	return &Root{
		NodeType: "root",
		Children: make([]Node, 0),
	}
}

// NewElement creates a new HAST element node
func NewElement(tagName string) *Element {
	return &Element{
		NodeType:   "element",
		TagName:    tagName,
		Properties: make(map[string]any),
		Children:   make([]Node, 0),
	}
}

// NewText creates a new HAST text node
func NewText(value string) *Text {
	return &Text{
		NodeType: "text",
		Value:    value,
	}
}

// NewComment creates a new HAST comment node
func NewComment(value string) *Comment {
	return &Comment{
		NodeType: "comment",
		Value:    value,
	}
}

// AddChild adds a child node to a parent node
func AddChild(parent Node, child Node) error {
	switch p := parent.(type) {
	case *Root:
		p.Children = append(p.Children, child)
	case *Element:
		p.Children = append(p.Children, child)
	default:
		return fmt.Errorf("cannot add child to node of type %s", parent.Type())
	}
	return nil
}

// SetProperty sets a property on an element
func (e *Element) SetProperty(key string, value any) {
	if e.Properties == nil {
		e.Properties = make(map[string]any)
	}
	e.Properties[key] = value
}

// GetProperty gets a property from an element
func (e *Element) GetProperty(key string) (any, bool) {
	if e.Properties == nil {
		return nil, false
	}
	value, exists := e.Properties[key]
	return value, exists
}

// HasProperty checks if an element has a specific property
func (e *Element) HasProperty(key string) bool {
	_, exists := e.GetProperty(key)
	return exists
}

// Custom JSON marshaling to handle the Node interface properly
func (r *Root) MarshalJSON() ([]byte, error) {
	type Alias Root
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "root",
		Alias: (*Alias)(r),
	})
}

func (e *Element) MarshalJSON() ([]byte, error) {
	type Alias Element
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "element",
		Alias: (*Alias)(e),
	})
}

func (t *Text) MarshalJSON() ([]byte, error) {
	type Alias Text
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "text",
		Alias: (*Alias)(t),
	})
}

func (c *Comment) MarshalJSON() ([]byte, error) {
	type Alias Comment
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "comment",
		Alias: (*Alias)(c),
	})
}
