package writers

import (
	"testing"
)

func TestNewStructuralStack(t *testing.T) {
	stack := NewStructuralStack()

	if stack == nil {
		t.Error("NewStructuralStack should return non-nil stack")
	}
	if stack.elements == nil {
		t.Error("Stack should have initialized elements slice")
	}
	if stack.Depth() != 0 {
		t.Error("New stack should have depth 0")
	}
	if !stack.IsEmpty() {
		t.Error("New stack should be empty")
	}
}

func TestStructuralStack_PushPop(t *testing.T) {
	stack := NewStructuralStack()

	// Push elements
	stack.Push(StackTypeList, 1, "test data")
	stack.Push(StackTypeTable, 0, nil)

	if stack.Depth() != 2 {
		t.Errorf("Expected depth 2, got %d", stack.Depth())
	}
	if stack.IsEmpty() {
		t.Error("Stack should not be empty after pushing elements")
	}

	// Pop elements
	element, ok := stack.Pop()
	if !ok {
		t.Error("Pop should return true for non-empty stack")
	}
	if element.Type != StackTypeTable {
		t.Errorf("Expected StackTypeTable, got %v", element.Type)
	}
	if element.Level != 0 {
		t.Errorf("Expected level 0, got %d", element.Level)
	}

	element, ok = stack.Pop()
	if !ok {
		t.Error("Pop should return true for non-empty stack")
	}
	if element.Type != StackTypeList {
		t.Errorf("Expected StackTypeList, got %v", element.Type)
	}
	if element.Level != 1 {
		t.Errorf("Expected level 1, got %d", element.Level)
	}
	if element.Data.(string) != "test data" {
		t.Errorf("Expected 'test data', got %v", element.Data)
	}

	// Pop from empty stack
	_, ok = stack.Pop()
	if ok {
		t.Error("Pop should return false for empty stack")
	}
}

func TestStructuralStack_Peek(t *testing.T) {
	stack := NewStructuralStack()

	// Peek empty stack
	_, ok := stack.Peek()
	if ok {
		t.Error("Peek should return false for empty stack")
	}

	// Push and peek
	stack.Push(StackTypeFormatting, 0, "formatting data")
	element, ok := stack.Peek()
	if !ok {
		t.Error("Peek should return true for non-empty stack")
	}
	if element.Type != StackTypeFormatting {
		t.Errorf("Expected StackTypeFormatting, got %v", element.Type)
	}
	if element.Data.(string) != "formatting data" {
		t.Errorf("Expected 'formatting data', got %v", element.Data)
	}

	// Verify peek doesn't modify stack
	if stack.Depth() != 1 {
		t.Error("Peek should not modify stack depth")
	}
}

func TestStructuralStack_Clear(t *testing.T) {
	stack := NewStructuralStack()

	// Add elements
	stack.Push(StackTypeList, 0, nil)
	stack.Push(StackTypeTable, 0, nil)
	stack.Push(StackTypeFormatting, 0, nil)

	if stack.Depth() != 3 {
		t.Errorf("Expected depth 3, got %d", stack.Depth())
	}

	// Clear
	stack.Clear()

	if stack.Depth() != 0 {
		t.Errorf("Expected depth 0 after clear, got %d", stack.Depth())
	}
	if !stack.IsEmpty() {
		t.Error("Stack should be empty after clear")
	}
}

func TestStructuralStack_PushList(t *testing.T) {
	stack := NewStructuralStack()

	stack.PushList(2, true, "custom-marker")

	element, ok := stack.Peek()
	if !ok {
		t.Error("Stack should not be empty after PushList")
	}
	if element.Type != StackTypeList {
		t.Errorf("Expected StackTypeList, got %v", element.Type)
	}
	if element.Level != 2 {
		t.Errorf("Expected level 2, got %d", element.Level)
	}

	listData, ok := element.Data.(*ListStackData)
	if !ok {
		t.Error("Element data should be ListStackData")
	}
	if !listData.Ordered {
		t.Error("List should be ordered")
	}
	if listData.ItemCount != 0 {
		t.Errorf("Expected item count 0, got %d", listData.ItemCount)
	}
	if listData.MarkerType != "custom-marker" {
		t.Errorf("Expected marker type 'custom-marker', got %q", listData.MarkerType)
	}
}

func TestStructuralStack_PushTable(t *testing.T) {
	stack := NewStructuralStack()

	stack.PushTable(3, 5, true)

	element, ok := stack.Peek()
	if !ok {
		t.Error("Stack should not be empty after PushTable")
	}
	if element.Type != StackTypeTable {
		t.Errorf("Expected StackTypeTable, got %v", element.Type)
	}

	tableData, ok := element.Data.(*TableStackData)
	if !ok {
		t.Error("Element data should be TableStackData")
	}
	if tableData.Columns != 3 {
		t.Errorf("Expected 3 columns, got %d", tableData.Columns)
	}
	if tableData.Rows != 5 {
		t.Errorf("Expected 5 rows, got %d", tableData.Rows)
	}
	if tableData.CurrentRow != 0 {
		t.Errorf("Expected current row 0, got %d", tableData.CurrentRow)
	}
	if tableData.CurrentCell != 0 {
		t.Errorf("Expected current cell 0, got %d", tableData.CurrentCell)
	}
	if !tableData.Headers {
		t.Error("Table should have headers")
	}
}

func TestStructuralStack_PushTableRow(t *testing.T) {
	stack := NewStructuralStack()

	// Push table first
	stack.PushTable(2, 3, false)
	stack.PushTableRow()

	if stack.Depth() != 2 {
		t.Errorf("Expected depth 2, got %d", stack.Depth())
	}

	// Check table row element
	element, ok := stack.Peek()
	if !ok {
		t.Error("Stack should not be empty")
	}
	if element.Type != StackTypeTableRow {
		t.Errorf("Expected StackTypeTableRow, got %v", element.Type)
	}

	// Check that parent table's current row was updated
	stack.Pop() // Remove table row to access table
	tableElement, ok := stack.Peek()
	if !ok {
		t.Error("Table should still be on stack")
	}
	tableData, ok := tableElement.Data.(*TableStackData)
	if !ok {
		t.Error("Element data should be TableStackData")
	}
	if tableData.CurrentRow != 1 {
		t.Errorf("Expected current row 1, got %d", tableData.CurrentRow)
	}
	if tableData.CurrentCell != 0 {
		t.Errorf("Expected current cell reset to 0, got %d", tableData.CurrentCell)
	}
}

func TestStructuralStack_PushTableCell(t *testing.T) {
	stack := NewStructuralStack()

	// Push table and row first
	stack.PushTable(2, 3, false)
	stack.PushTableRow()
	stack.PushTableCell()

	if stack.Depth() != 3 {
		t.Errorf("Expected depth 3, got %d", stack.Depth())
	}

	// Check table cell element
	element, ok := stack.Peek()
	if !ok {
		t.Error("Stack should not be empty")
	}
	if element.Type != StackTypeTableCell {
		t.Errorf("Expected StackTypeTableCell, got %v", element.Type)
	}

	// Pop cell and row to check table
	stack.Pop()
	stack.Pop()
	tableElement, ok := stack.Peek()
	if !ok {
		t.Error("Table should still be on stack")
	}
	tableData, ok := tableElement.Data.(*TableStackData)
	if !ok {
		t.Error("Element data should be TableStackData")
	}
	if tableData.CurrentCell != 1 {
		t.Errorf("Expected current cell 1, got %d", tableData.CurrentCell)
	}
}

func TestStructuralStack_PushFormatting(t *testing.T) {
	stack := NewStructuralStack()

	stack.PushFormatting(0x0F, "https://example.com", "<strong>", "</strong>")

	element, ok := stack.Peek()
	if !ok {
		t.Error("Stack should not be empty after PushFormatting")
	}
	if element.Type != StackTypeFormatting {
		t.Errorf("Expected StackTypeFormatting, got %v", element.Type)
	}

	formatData, ok := element.Data.(*FormattingStackData)
	if !ok {
		t.Error("Element data should be FormattingStackData")
	}
	if formatData.StyleFlags != 0x0F {
		t.Errorf("Expected style flags 0x0F, got 0x%X", formatData.StyleFlags)
	}
	if formatData.LinkURL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got %q", formatData.LinkURL)
	}
	if formatData.StartTag != "<strong>" {
		t.Errorf("Expected start tag '<strong>', got %q", formatData.StartTag)
	}
	if formatData.EndTag != "</strong>" {
		t.Errorf("Expected end tag '</strong>', got %q", formatData.EndTag)
	}
}

func TestStructuralStack_GetCurrentList(t *testing.T) {
	stack := NewStructuralStack()

	// No list on stack
	_, found := stack.GetCurrentList()
	if found {
		t.Error("Should not find list when none exists")
	}

	// Push non-list element
	stack.Push(StackTypeTable, 0, nil)
	_, found = stack.GetCurrentList()
	if found {
		t.Error("Should not find list when only table exists")
	}

	// Push list
	stack.PushList(1, false, "bullet")
	listData, found := stack.GetCurrentList()
	if !found {
		t.Error("Should find list when it exists")
	}
	if listData == nil {
		t.Error("List data should not be nil")
	}
	if listData.Ordered {
		t.Error("List should be unordered")
	}
	if listData.MarkerType != "bullet" {
		t.Errorf("Expected marker type 'bullet', got %q", listData.MarkerType)
	}

	// Push formatting on top
	stack.PushFormatting(0x01, "", "", "")
	listData, found = stack.GetCurrentList()
	if !found {
		t.Error("Should still find list even with formatting on top")
	}
	if listData.MarkerType != "bullet" {
		t.Error("Should find the same list")
	}
}

func TestStructuralStack_GetCurrentTable(t *testing.T) {
	stack := NewStructuralStack()

	// No table on stack
	_, found := stack.GetCurrentTable()
	if found {
		t.Error("Should not find table when none exists")
	}

	// Push list
	stack.PushList(0, false, "")
	_, found = stack.GetCurrentTable()
	if found {
		t.Error("Should not find table when only list exists")
	}

	// Push table
	stack.PushTable(4, 6, true)
	tableData, found := stack.GetCurrentTable()
	if !found {
		t.Error("Should find table when it exists")
	}
	if tableData == nil {
		t.Error("Table data should not be nil")
	}
	if tableData.Columns != 4 {
		t.Errorf("Expected 4 columns, got %d", tableData.Columns)
	}
	if tableData.Rows != 6 {
		t.Errorf("Expected 6 rows, got %d", tableData.Rows)
	}

	// Push table row on top
	stack.PushTableRow()
	tableData, found = stack.GetCurrentTable()
	if !found {
		t.Error("Should still find table even with row on top")
	}
	if tableData.Columns != 4 {
		t.Error("Should find the same table")
	}
}

func TestStructuralStack_IncrementListItem(t *testing.T) {
	stack := NewStructuralStack()

	// No list on stack
	err := stack.IncrementListItem()
	if err == nil {
		t.Error("Should return error when no list exists")
	}

	// Push list
	stack.PushList(0, true, "")
	err = stack.IncrementListItem()
	if err != nil {
		t.Errorf("Should not return error when list exists: %v", err)
	}

	listData, found := stack.GetCurrentList()
	if !found {
		t.Error("Should find list")
	}
	if listData.ItemCount != 1 {
		t.Errorf("Expected item count 1, got %d", listData.ItemCount)
	}

	// Increment again
	err = stack.IncrementListItem()
	if err != nil {
		t.Errorf("Should not return error: %v", err)
	}
	if listData.ItemCount != 2 {
		t.Errorf("Expected item count 2, got %d", listData.ItemCount)
	}
}

func TestStructuralStack_GetListLevel(t *testing.T) {
	stack := NewStructuralStack()

	// No lists
	level := stack.GetListLevel()
	if level != 0 {
		t.Errorf("Expected list level 0, got %d", level)
	}

	// One list
	stack.PushList(0, false, "")
	level = stack.GetListLevel()
	if level != 1 {
		t.Errorf("Expected list level 1, got %d", level)
	}

	// Nested lists
	stack.Push(StackTypeFormatting, 0, nil) // Non-list element
	stack.PushList(1, false, "")
	stack.PushList(2, false, "")
	level = stack.GetListLevel()
	if level != 3 {
		t.Errorf("Expected list level 3, got %d", level)
	}
}

func TestStructuralStack_GetTableContext(t *testing.T) {
	stack := NewStructuralStack()

	// No table
	row, cell, cols, inTable := stack.GetTableContext()
	if inTable {
		t.Error("Should not be in table context")
	}
	if row != 0 || cell != 0 || cols != 0 {
		t.Error("All values should be 0 when not in table")
	}

	// Push table
	stack.PushTable(3, 5, false)
	row, cell, cols, inTable = stack.GetTableContext()
	if !inTable {
		t.Error("Should be in table context")
	}
	if cols != 3 {
		t.Errorf("Expected 3 columns, got %d", cols)
	}
	if row != 0 {
		t.Errorf("Expected row 0, got %d", row)
	}
	if cell != 0 {
		t.Errorf("Expected cell 0, got %d", cell)
	}

	// Add row and cells
	stack.PushTableRow()
	stack.PushTableCell()
	stack.Pop() // Remove cell to update table
	stack.PushTableCell()

	row, cell, cols, inTable = stack.GetTableContext()
	if !inTable {
		t.Error("Should still be in table context")
	}
	if row != 1 {
		t.Errorf("Expected row 1, got %d", row)
	}
	if cell != 2 {
		t.Errorf("Expected cell 2, got %d", cell)
	}
}

func TestStructuralStack_IsInContext(t *testing.T) {
	stack := NewStructuralStack()

	// Empty stack
	if stack.IsInContext(StackTypeList) {
		t.Error("Should not be in list context")
	}
	if stack.IsInContext(StackTypeTable) {
		t.Error("Should not be in table context")
	}

	// Push list
	stack.PushList(0, false, "")
	if !stack.IsInContext(StackTypeList) {
		t.Error("Should be in list context")
	}
	if stack.IsInContext(StackTypeTable) {
		t.Error("Should not be in table context")
	}

	// Push table
	stack.PushTable(2, 2, false)
	if !stack.IsInContext(StackTypeList) {
		t.Error("Should still be in list context")
	}
	if !stack.IsInContext(StackTypeTable) {
		t.Error("Should be in table context")
	}

	// Push formatting
	stack.PushFormatting(0x01, "", "", "")
	if !stack.IsInContext(StackTypeList) {
		t.Error("Should still be in list context")
	}
	if !stack.IsInContext(StackTypeTable) {
		t.Error("Should still be in table context")
	}
	if !stack.IsInContext(StackTypeFormatting) {
		t.Error("Should be in formatting context")
	}
}

func TestStructuralStack_ValidateStack(t *testing.T) {
	stack := NewStructuralStack()

	// Empty stack is valid
	err := stack.ValidateStack()
	if err != nil {
		t.Errorf("Empty stack should be valid: %v", err)
	}

	// Push elements without popping
	stack.Push(StackTypeList, 0, nil)
	stack.Push(StackTypeTable, 0, nil)
	stack.Push(StackTypeFormatting, 0, nil)

	err = stack.ValidateStack()
	if err == nil {
		t.Error("Stack with unclosed elements should be invalid")
	}
	if err.Error() != "unclosed structural elements: [list table formatting]" {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Pop all elements
	stack.Pop()
	stack.Pop()
	stack.Pop()

	err = stack.ValidateStack()
	if err != nil {
		t.Errorf("Balanced stack should be valid: %v", err)
	}
}

func TestStructuralStack_ComplexNesting(t *testing.T) {
	stack := NewStructuralStack()

	// Simulate complex nested structure:
	// List -> Table -> Row -> Cell -> Formatting
	stack.PushList(0, false, "bullet")
	stack.PushTable(2, 3, true)
	stack.PushTableRow()
	stack.PushTableCell()
	stack.PushFormatting(0x03, "https://example.com", "<em>", "</em>")

	if stack.Depth() != 5 {
		t.Errorf("Expected depth 5, got %d", stack.Depth())
	}

	// Verify contexts
	if !stack.IsInContext(StackTypeList) {
		t.Error("Should be in list context")
	}
	if !stack.IsInContext(StackTypeTable) {
		t.Error("Should be in table context")
	}
	if !stack.IsInContext(StackTypeTableRow) {
		t.Error("Should be in table row context")
	}
	if !stack.IsInContext(StackTypeTableCell) {
		t.Error("Should be in table cell context")
	}
	if !stack.IsInContext(StackTypeFormatting) {
		t.Error("Should be in formatting context")
	}

	// Verify list level
	if stack.GetListLevel() != 1 {
		t.Errorf("Expected list level 1, got %d", stack.GetListLevel())
	}

	// Verify table context
	row, cell, cols, inTable := stack.GetTableContext()
	if !inTable {
		t.Error("Should be in table context")
	}
	if cols != 2 {
		t.Errorf("Expected 2 columns, got %d", cols)
	}
	if row != 1 {
		t.Errorf("Expected row 1, got %d", row)
	}
	if cell != 1 {
		t.Errorf("Expected cell 1, got %d", cell)
	}

	// Pop all elements in correct order
	element, ok := stack.Pop()
	if !ok || element.Type != StackTypeFormatting {
		t.Error("Should pop formatting first")
	}
	element, ok = stack.Pop()
	if !ok || element.Type != StackTypeTableCell {
		t.Error("Should pop table cell second")
	}
	element, ok = stack.Pop()
	if !ok || element.Type != StackTypeTableRow {
		t.Error("Should pop table row third")
	}
	element, ok = stack.Pop()
	if !ok || element.Type != StackTypeTable {
		t.Error("Should pop table fourth")
	}
	element, ok = stack.Pop()
	if !ok || element.Type != StackTypeList {
		t.Error("Should pop list last")
	}

	if !stack.IsEmpty() {
		t.Error("Stack should be empty after popping all elements")
	}
}

// Test edge cases and error conditions
func TestStructuralStack_EdgeCases(t *testing.T) {
	stack := NewStructuralStack()

	// Test with nil data
	stack.Push(StackTypeList, 0, nil)
	element, ok := stack.Pop()
	if !ok {
		t.Error("Should be able to pop element with nil data")
	}
	if element.Data != nil {
		t.Error("Data should be nil")
	}

	// Test multiple table operations without table
	stack.PushTableRow()
	stack.PushTableCell()
	if stack.Depth() != 2 {
		t.Error("Should be able to push table elements without parent table")
	}

	// Test GetCurrentTable with table row but no table
	_, found := stack.GetCurrentTable()
	if found {
		t.Error("Should not find table when only row exists")
	}

	// Test increment list item with table context
	err := stack.IncrementListItem()
	if err == nil {
		t.Error("Should return error when not in list context")
	}
}

// Performance test
func BenchmarkStructuralStack_PushPop(b *testing.B) {
	stack := NewStructuralStack()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.Push(StackTypeList, i%10, "test data")
		if i%100 == 0 {
			// Occasionally pop to prevent unlimited growth
			stack.Pop()
		}
	}
}

func BenchmarkStructuralStack_GetCurrentList(b *testing.B) {
	stack := NewStructuralStack()

	// Set up a deep stack with list at bottom
	stack.PushList(0, false, "test")
	for i := 0; i < 100; i++ {
		stack.Push(StackTypeFormatting, 0, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = stack.GetCurrentList()
	}
}

func BenchmarkStructuralStack_IsInContext(b *testing.B) {
	stack := NewStructuralStack()

	// Set up a deep stack
	for i := 0; i < 100; i++ {
		stack.Push(StackType(i%5), i, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stack.IsInContext(StackTypeList)
	}
}
