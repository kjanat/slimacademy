package writers

import (
	"fmt"
)

// StackElement represents a structural element in the document hierarchy
type StackElement struct {
	Type  StackType
	Level int
	Data  any // Format-specific data
}

// StackType represents the type of structural element
type StackType uint8

const (
	StackTypeList StackType = iota
	StackTypeTable
	StackTypeTableRow
	StackTypeTableCell
	StackTypeFormatting
)

// StructuralStack manages nested document structures (lists, tables, formatting)
type StructuralStack struct {
	elements []StackElement
}

// NewStructuralStack returns a new StructuralStack initialized with capacity for common nesting depth.
func NewStructuralStack() *StructuralStack {
	return &StructuralStack{
		elements: make([]StackElement, 0, 8), // Pre-allocate for common depth
	}
}

// Push adds an element to the stack
func (s *StructuralStack) Push(elementType StackType, level int, data any) {
	s.elements = append(s.elements, StackElement{
		Type:  elementType,
		Level: level,
		Data:  data,
	})
}

// Pop removes and returns the top element
func (s *StructuralStack) Pop() (StackElement, bool) {
	if len(s.elements) == 0 {
		return StackElement{}, false
	}

	top := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return top, true
}

// Peek returns the top element without removing it
func (s *StructuralStack) Peek() (StackElement, bool) {
	if len(s.elements) == 0 {
		return StackElement{}, false
	}
	return s.elements[len(s.elements)-1], true
}

// Depth returns the current stack depth
func (s *StructuralStack) Depth() int {
	return len(s.elements)
}

// IsEmpty returns true if the stack is empty
func (s *StructuralStack) IsEmpty() bool {
	return len(s.elements) == 0
}

// Clear removes all elements from the stack
func (s *StructuralStack) Clear() {
	s.elements = s.elements[:0]
}

// ListStackData contains list-specific information
type ListStackData struct {
	Ordered    bool
	ItemCount  int
	MarkerType string // For custom list markers
}

// TableStackData contains table-specific information
type TableStackData struct {
	Columns     int
	Rows        int
	CurrentRow  int
	CurrentCell int
	Headers     bool // Whether first row is headers
}

// FormattingStackData contains formatting-specific information
type FormattingStackData struct {
	StyleFlags uint16
	LinkURL    string
	StartTag   string // For format-specific opening tags
	EndTag     string // For format-specific closing tags
}

// Helper methods for common stack operations

// PushList adds a list to the stack
func (s *StructuralStack) PushList(level int, ordered bool, markerType string) {
	data := &ListStackData{
		Ordered:    ordered,
		ItemCount:  0,
		MarkerType: markerType,
	}
	s.Push(StackTypeList, level, data)
}

// PushTable adds a table to the stack
func (s *StructuralStack) PushTable(columns, rows int, hasHeaders bool) {
	data := &TableStackData{
		Columns:     columns,
		Rows:        rows,
		CurrentRow:  0,
		CurrentCell: 0,
		Headers:     hasHeaders,
	}
	s.Push(StackTypeTable, 0, data)
}

// PushTableRow adds a table row to the stack
func (s *StructuralStack) PushTableRow() {
	s.Push(StackTypeTableRow, 0, nil)

	// Update parent table's current row
	if s.Depth() >= 2 {
		parentIdx := len(s.elements) - 2
		if s.elements[parentIdx].Type == StackTypeTable {
			if tableData, ok := s.elements[parentIdx].Data.(*TableStackData); ok {
				tableData.CurrentRow++
				tableData.CurrentCell = 0
			}
		}
	}
}

// PushTableCell adds a table cell to the stack
func (s *StructuralStack) PushTableCell() {
	s.Push(StackTypeTableCell, 0, nil)

	// Update parent table's current cell
	for i := len(s.elements) - 2; i >= 0; i-- {
		if s.elements[i].Type == StackTypeTable {
			if tableData, ok := s.elements[i].Data.(*TableStackData); ok {
				tableData.CurrentCell++
			}
			break
		}
	}
}

// PushFormatting adds formatting to the stack
func (s *StructuralStack) PushFormatting(styleFlags uint16, linkURL, startTag, endTag string) {
	data := &FormattingStackData{
		StyleFlags: styleFlags,
		LinkURL:    linkURL,
		StartTag:   startTag,
		EndTag:     endTag,
	}
	s.Push(StackTypeFormatting, 0, data)
}

// GetCurrentList returns the current list data if we're in a list
func (s *StructuralStack) GetCurrentList() (*ListStackData, bool) {
	for i := len(s.elements) - 1; i >= 0; i-- {
		if s.elements[i].Type == StackTypeList {
			if data, ok := s.elements[i].Data.(*ListStackData); ok {
				return data, true
			}
		}
	}
	return nil, false
}

// GetCurrentTable returns the current table data if we're in a table
func (s *StructuralStack) GetCurrentTable() (*TableStackData, bool) {
	for i := len(s.elements) - 1; i >= 0; i-- {
		if s.elements[i].Type == StackTypeTable {
			if data, ok := s.elements[i].Data.(*TableStackData); ok {
				return data, true
			}
		}
	}
	return nil, false
}

// IncrementListItem increments the item count for the current list
func (s *StructuralStack) IncrementListItem() error {
	listData, found := s.GetCurrentList()
	if !found {
		return fmt.Errorf("not currently in a list")
	}
	listData.ItemCount++
	return nil
}

// GetListLevel returns the current list nesting level
func (s *StructuralStack) GetListLevel() int {
	level := 0
	for _, element := range s.elements {
		if element.Type == StackTypeList {
			level++
		}
	}
	return level
}

// GetTableContext returns current table position information
func (s *StructuralStack) GetTableContext() (row, cell, totalCols int, inTable bool) {
	tableData, found := s.GetCurrentTable()
	if !found {
		return 0, 0, 0, false
	}
	return tableData.CurrentRow, tableData.CurrentCell, tableData.Columns, true
}

// IsInContext checks if we're currently in a specific structural context
func (s *StructuralStack) IsInContext(contextType StackType) bool {
	for _, element := range s.elements {
		if element.Type == contextType {
			return true
		}
	}
	return false
}

// ValidateStack checks for balanced open/close operations
func (s *StructuralStack) ValidateStack() error {
	if !s.IsEmpty() {
		var unclosed []string
		for _, element := range s.elements {
			var typeName string
			switch element.Type {
			case StackTypeList:
				typeName = "list"
			case StackTypeTable:
				typeName = "table"
			case StackTypeTableRow:
				typeName = "table row"
			case StackTypeTableCell:
				typeName = "table cell"
			case StackTypeFormatting:
				typeName = "formatting"
			default:
				typeName = "unknown"
			}
			unclosed = append(unclosed, typeName)
		}
		return fmt.Errorf("unclosed structural elements: %v", unclosed)
	}
	return nil
}
