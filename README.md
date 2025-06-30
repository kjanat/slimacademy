# Slim Academy Book Transformer

This project transforms JSON data from Slim Academy into an internal data model. From this intermediate representation, the content can be exported to various formats like Markdown, HTML, or ePub.

## Project Structure

Each book folder contains:

- **Metadata JSON** (e.g., `3631.json`): Book information, images, dates, purchase status
- **chapters.json**: Hierarchical chapter structure with nested subchapters  
- **content.json**: Google Docs-like rich content with paragraphs, styling, and formatting
- **list-notes.json**: Additional notes (typically empty)

## Installation

```bash
git clone https://github.com/kjanat/slimacademy.git
cd slimacademy
go mod tidy
go build -o bin/transformer ./cmd/transformer
```

## Usage

### List Available Books

```bash
./bin/transformer --list
```

### Transform a Specific Book

```bash
./bin/transformer --book "Station B3 Deel 1" --format markdown --output ./output
```

### Transform All Books

```bash
./bin/transformer --format markdown --output ./output
```

### Available Formats

- `markdown`: Export as Markdown (.md)
- `html`: Export as HTML (.html)
- `epub`: Export as ePub-compatible HTML (.epub.html)

### Command Line Options

- `--input`: Input directory containing book folders (default: current directory)
- `--output`: Output directory for exported files (default: ./output)
- `--format`: Export format (markdown, html, epub)
- `--book`: Specific book directory to transform (optional)
- `--list`: List all available books

## Architecture

The project follows a clean architecture pattern:

```plaintext
slimacademy/
├── cmd/transformer/         # CLI entry point
├── internal/
│   ├── models/              # Data structures for books
│   ├── parser/              # JSON parsing logic
│   ├── transformer/         # Core transformation engine
│   └── exporters/           # Format-specific exporters
├── pkg/exporters/           # Export interfaces
```

## Features

- **Hierarchical content parsing**: Preserves chapter structure and nested content
- **Rich text formatting**: Supports bold, italic, underline, links
- **Multiple export formats**: Markdown, HTML, and ePub-ready output
- **Table of contents generation**: Automatic TOC based on chapter structure
- **Clean text processing**: Removes formatting artifacts and normalizes content

## Example Output

The transformer generates clean, structured documents with:

- Proper heading hierarchy
- Table of contents with anchor links
- Formatted text with styling preserved
- Cross-references and links maintained
