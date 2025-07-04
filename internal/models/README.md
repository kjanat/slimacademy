# Internal Models Directory Structure

This directory contains the core data structures and types used in the document transformation system.

I deleted the previous file structure and replaced it with a more organized and modular structure.

```tree
internal/models/
├── book.go        # Book and BookImage types
├── chapter.go     # Chapter type
├── document.go    # Document, Body, and structural elements
├── paragraph.go   # Paragraph and text-related elements
├── style.go       # All styling types (TextStyle, ParagraphStyle, etc.)
├── table.go       # Table-related types
├── objects.go     # Inline and positioned objects
├── list.go        # List-related types
├── common.go      # Common types like Dimension and Size
└── content.go     # Content union type and unmarshal helper
```
