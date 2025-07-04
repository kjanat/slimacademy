# SlimAcademy Document Transformation System - User Guide

Welcome to SlimAcademy! This guide will help you transform your academic documents into various formats with ease.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Installation](#installation)
3. [Basic Usage](#basic-usage)
4. [Tutorial: Your First Conversion](#tutorial-your-first-conversion)
5. [Command Reference](#command-reference)
6. [Working with Different Formats](#working-with-different-formats)
7. [Advanced Features](#advanced-features)
8. [Troubleshooting](#troubleshooting)
9. [FAQ](#faq)

## Getting Started

SlimAcademy is a powerful document transformation system designed specifically for academic content. It can convert your documents between multiple formats while preserving formatting, structure, and academic metadata.

### What Can SlimAcademy Do?

- 📚 Convert academic documents to HTML, Markdown, LaTeX, and EPUB
- 🎨 Preserve complex formatting, tables, and lists
- 📊 Handle academic metadata (citations, references, etc.)
- ⚡ Process large documents efficiently with streaming
- 🔄 Batch convert multiple documents at once

## Installation

### Prerequisites

- Go 1.21 or higher
- Git (for cloning the repository)

### Step 1: Clone the Repository

```bash
git clone https://github.com/yourusername/slimacademy.git
cd slimacademy
```

### Step 2: Build the Application

```bash
go build -o slim ./cmd/slim
```

This creates an executable named `slim` in your current directory.

### Step 3: Verify Installation

```bash
./slim --help
```

You should see the help message with available commands.

## Basic Usage

The SlimAcademy CLI provides several commands for different tasks:

- `convert` - Transform documents to different formats
- `list` - List available documents
- `check` - Validate document structure
- `fetch` - Download documents from SlimAcademy (requires authentication)

## Tutorial: Your First Conversion

Let's walk through converting your first document!

### Step 1: List Available Documents

First, let's see what documents are available:

```bash
./slim list source/
```

This will show all documents in the source directory:

```
Available books in source/:
- Station B3 Deel 1 (ID: 3917)
- Station B3 Deel 2 (ID: 3918)
- Station B3 Deel 3 (ID: 3919)
[...]
```

### Step 2: Convert a Single Document

Let's convert "Station B3 Deel 1" to HTML:

```bash
./slim convert --format html "Station B3 Deel 1"
```

The converted file will be saved as `Station B3 Deel 1.html` in the current directory.

### Step 3: Convert to Multiple Formats

You can convert to multiple formats at once:

```bash
./slim convert --formats "html,markdown,latex" "Station B3 Deel 1"
```

This creates three files:
- `Station B3 Deel 1.html`
- `Station B3 Deel 1.md`
- `Station B3 Deel 1.tex`

### Step 4: Specify Output Directory

To organize your converted files:

```bash
./slim convert --formats "html,markdown" --output converted/ "Station B3 Deel 1"
```

Files will be saved in the `converted/` directory.

### Step 5: Convert All Documents

To convert all documents at once:

```bash
./slim convert --all --output output/
```

This creates a ZIP file containing all documents in all formats!

## Command Reference

### convert Command

Transform documents to different formats.

```bash
./slim convert [options] [book-path]
```

**Options:**
- `--all` - Convert all available books
- `--format FORMAT` - Convert to a single format (html, markdown, latex, epub)
- `--formats "FORMAT1,FORMAT2"` - Convert to multiple formats
- `--output PATH` - Output directory (default: current directory)
- `--config PATH` - Path to configuration file

**Examples:**

```bash
# Convert specific book to HTML
./slim convert --format html "My Book"

# Convert all books to all formats
./slim convert --all

# Convert to specific formats with custom output
./slim convert --formats "html,epub" --output ~/Documents/books/ "My Book"
```

### list Command

List available documents.

```bash
./slim list [source-directory]
```

**Example:**

```bash
./slim list source/
```

### check Command

Validate document structure and report issues.

```bash
./slim check [book-path]
```

**Example:**

```bash
./slim check "Station B3 Deel 1"
```

### fetch Command

Download documents from SlimAcademy (requires authentication).

```bash
./slim fetch [options]
```

**Options:**
- `--all` - Fetch all available books
- `--book-id ID` - Fetch specific book by ID
- `--login` - Login to SlimAcademy
- `--output PATH` - Where to save fetched documents
- `--clean` - Remove existing source data before fetching

**First-time Setup:**

1. Create a `.env` file with your credentials:
```
USERNAME=your.email@example.com
PASSWORD=yourpassword
```

2. Login:
```bash
./slim fetch --login
```

3. Fetch all books:
```bash
./slim fetch --all
```

## Working with Different Formats

### HTML Output

HTML files include:
- Responsive design with academic styling
- Table of contents with navigation
- Proper heading hierarchy
- Syntax highlighting for code blocks
- Mathematical equation rendering

**Best for:** Web viewing, online sharing, printing

### Markdown Output

Markdown files preserve:
- Document structure and headings
- Lists and tables
- Links and references
- Code blocks with language hints
- Basic text formatting

**Best for:** Documentation, GitHub, note-taking apps

### LaTeX Output

LaTeX files include:
- Academic document class setup
- Proper sectioning commands
- Table and figure environments
- Bibliography support
- Mathematical equations

**Best for:** Academic papers, thesis writing, professional publishing

### EPUB Output

EPUB files feature:
- E-reader compatibility
- Chapter navigation
- Responsive text flow
- Embedded images
- Metadata preservation

**Best for:** E-readers, mobile reading, digital libraries

## Advanced Features

### Batch Processing

Convert multiple specific books:

```bash
for book in "Book 1" "Book 2" "Book 3"; do
  ./slim convert --formats "html,epub" "$book"
done
```

### Custom Configuration

Create a configuration file for repeated use:

```yaml
# config.yaml
output_directory: ~/Documents/SlimAcademy/
formats:
  - html
  - markdown
  - epub
html_options:
  include_toc: true
  syntax_highlighting: true
```

Use with:
```bash
./slim convert --config config.yaml --all
```

### Parallel Processing

The system automatically uses parallel processing for:
- Multiple format conversions
- Batch document processing
- Large file streaming

## Troubleshooting

### Common Issues

#### "Book not found" Error

**Problem:** The specified book name doesn't match exactly.

**Solution:** Use `./slim list` to see exact book names, including special characters.

#### Large File Processing

**Problem:** Processing hangs with very large documents.

**Solution:** The system automatically uses streaming for files over 10MB. If issues persist:
- Check available disk space
- Monitor memory usage
- Consider converting formats individually

#### Authentication Errors

**Problem:** Cannot fetch documents from SlimAcademy.

**Solutions:**
1. Check `.env` file exists and has correct credentials
2. Try logging in again: `./slim fetch --login`
3. Ensure credentials use email format for username

#### Output File Conflicts

**Problem:** "File already exists" errors.

**Solution:** Either:
- Remove existing files manually
- Use a different output directory
- Add timestamp to filenames

### Debug Mode

For detailed error information:

```bash
# Set debug environment variable
export DEBUG=1
./slim convert --format html "My Book"
```

## FAQ

### Q: Can I convert only specific chapters?

Currently, SlimAcademy converts complete documents. Chapter selection is planned for future releases.

### Q: How do I handle documents with images?

Images are automatically processed and embedded in HTML and EPUB formats. For Markdown and LaTeX, image references are preserved.

### Q: Can I customize the output styling?

Yes! Each format supports customization:
- HTML: Edit the embedded CSS
- LaTeX: Modify the preamble
- Markdown: Use pandoc for further processing

### Q: Is there a file size limit?

No hard limit, but documents over 10MB use streaming for efficiency. The largest tested document is 17MB.

### Q: Can I use SlimAcademy in my CI/CD pipeline?

Yes! SlimAcademy is designed for automation:
```bash
# Example GitHub Action step
- name: Convert documents
  run: |
    ./slim convert --all --output artifacts/
```

### Q: How do I update SlimAcademy?

```bash
git pull origin main
go build -o slim ./cmd/slim
```

## Need More Help?

- 📧 Report issues: [GitHub Issues](https://github.com/yourusername/slimacademy/issues)
- 📚 Read the code documentation: [CLAUDE.md](CLAUDE.md)
- 🤝 Contribute: See [CONTRIBUTING.md](CONTRIBUTING.md)

---

Happy document transforming! 🎓📚
