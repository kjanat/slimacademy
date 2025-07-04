# SlimAcademy Quick Reference

## Installation
```bash
git clone https://github.com/yourusername/slimacademy.git
cd slimacademy
go build -o slim ./cmd/slim
```

## Basic Commands

### List Documents
```bash
./slim list source/
```

### Convert Single Document
```bash
# To HTML
./slim convert --format html "Book Name"

# To multiple formats
./slim convert --formats "html,markdown,latex" "Book Name"

# With output directory
./slim convert --format html --output docs/ "Book Name"
```

### Convert All Documents
```bash
# Creates a ZIP with all formats
./slim convert --all

# Specific formats only
./slim convert --all --formats "html,epub"
```

### Validate Documents
```bash
./slim check "Book Name"
```

### Fetch from SlimAcademy
```bash
# First time setup - create .env file:
echo "USERNAME=your.email@example.com" > .env
echo "PASSWORD=yourpassword" >> .env

# Login
./slim fetch --login

# Fetch all books
./slim fetch --all

# Fetch specific book
./slim fetch --book-id 3917
```

## Format Options

| Format | Extension | Best For |
|--------|-----------|----------|
| HTML | .html | Web viewing, sharing |
| Markdown | .md | Documentation, notes |
| LaTeX | .tex | Academic papers |
| EPUB | .epub | E-readers |

## Common Options

- `--all` - Process all documents
- `--format FORMAT` - Single format output
- `--formats "F1,F2"` - Multiple formats
- `--output PATH` - Output directory
- `--config FILE` - Configuration file

## Tips

1. **Exact Names**: Book names must match exactly (use `list` command)
2. **Batch Processing**: Use `--all` for multiple documents
3. **Organization**: Use `--output` to organize converted files
4. **Performance**: Large files (>10MB) automatically use streaming

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Book not found | Check exact name with `list` |
| Auth fails | Verify .env file, re-login |
| File exists | Use different output directory |
| Large file hangs | Check disk space/memory |

## Debug Mode
```bash
export DEBUG=1
./slim convert --format html "Book Name"
```
