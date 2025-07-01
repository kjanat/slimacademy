package writers

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/streaming"
)

// init registers the EPUB writer with associated metadata for use in the writers package.
func init() {
	Register("epub", func() WriterV2 {
		return &EPUBWriterV2{
			stats: WriterStats{},
		}
	}, WriterMetadata{
		Name:        "EPUB",
		Extension:   ".epub",
		Description: "EPUB e-book format",
		MimeType:    "application/epub+zip",
	})
}

// EPUBWriter generates EPUB files using HTML content
type EPUBWriter struct {
	config         *config.EPUBConfig
	htmlWriter     *HTMLWriter
	zipWriter      *zip.Writer
	output         io.Writer
	title          string
	uuid           string
	chapters       []Chapter
	currentChapter *Chapter
	lastError      error
}

// Chapter represents a chapter in the EPUB
type Chapter struct {
	ID       string
	Title    string
	Filename string
	Content  string
}

// NewEPUBWriter returns a new EPUBWriter that writes an EPUB file to the specified output using default configuration.
func NewEPUBWriter(output io.Writer) *EPUBWriter {
	return NewEPUBWriterWithConfig(output, nil)
}

// NewEPUBWriterWithConfig returns a new EPUBWriter that writes to the specified output using the provided EPUB configuration.
// If the configuration is nil, a default configuration is used. The writer is initialized with a ZIP archive and a unique identifier.
func NewEPUBWriterWithConfig(output io.Writer, cfg *config.EPUBConfig) *EPUBWriter {
	if cfg == nil {
		cfg = config.DefaultEPUBConfig()
	}
	zipWriter := zip.NewWriter(output)
	return &EPUBWriter{
		config:     cfg,
		htmlWriter: NewHTMLWriterWithConfig(cfg.HTMLConfig),
		zipWriter:  zipWriter,
		output:     output,
		uuid:       generateUUID(),
		chapters:   make([]Chapter, 0),
	}
}

// Handle processes a single event
func (w *EPUBWriter) Handle(event streaming.Event) {
	switch event.Kind {
	case streaming.StartDoc:
		w.title = event.Title
		// Initialize the HTML writer
		w.htmlWriter.Reset()

	case streaming.StartHeading:
		// Create a new chapter for each heading
		if w.currentChapter != nil {
			// Finalize previous chapter
			w.currentChapter.Content = w.htmlWriter.Result()
			w.chapters = append(w.chapters, *w.currentChapter)
		}

		// Start new chapter
		w.currentChapter = &Chapter{
			ID:       event.AnchorID,
			Title:    event.HeadingText.Value(),
			Filename: w.config.GetChapterFilename(event.HeadingText.Value(), event.AnchorID),
		}
		w.htmlWriter.Reset()
		w.htmlWriter.Handle(streaming.Event{Kind: streaming.StartDoc, Title: event.HeadingText.Value()})

	case streaming.EndDoc:
		// Finalize last chapter
		if w.currentChapter != nil {
			w.htmlWriter.Handle(event)
			w.currentChapter.Content = w.htmlWriter.Result()
			w.chapters = append(w.chapters, *w.currentChapter)
		}

		// Generate EPUB files
		if err := w.generateEPUB(); err != nil {
			// Store error for later retrieval instead of printing
			w.lastError = err
		}

	default:
		// Forward all other events to HTML writer
		w.htmlWriter.Handle(event)
	}
}

// generateEPUB creates the EPUB file structure
func (w *EPUBWriter) generateEPUB() error {
	// Write mimetype file (must be first and uncompressed)
	mimeWriter, err := w.zipWriter.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store, // No compression
	})
	if err != nil {
		return err
	}
	if _, err := mimeWriter.Write([]byte("application/epub+zip")); err != nil {
		return err
	}

	// Write META-INF/container.xml
	if err := w.writeFile("META-INF/container.xml", w.getContainerXML()); err != nil {
		return err
	}

	// Write content.opf
	if err := w.writeFile("OEBPS/content.opf", w.getContentOPF()); err != nil {
		return err
	}

	// Write toc.ncx
	if err := w.writeFile("OEBPS/toc.ncx", w.getTocNCX()); err != nil {
		return err
	}

	// Write chapter files
	for _, chapter := range w.chapters {
		if err := w.writeFile(fmt.Sprintf("OEBPS/%s", chapter.Filename), chapter.Content); err != nil {
			return err
		}
	}

	// Write CSS
	if err := w.writeFile("OEBPS/styles.css", w.config.GetDefaultCSS()); err != nil {
		return err
	}

	return w.zipWriter.Close()
}

// writeFile writes a file to the ZIP archive
func (w *EPUBWriter) writeFile(filename, content string) error {
	writer, err := w.zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(content))
	return err
}

// getContainerXML returns the container.xml content
func (w *EPUBWriter) getContainerXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
}

// getContentOPF returns the content.opf content
func (w *EPUBWriter) getContentOPF() string {
	var manifest strings.Builder
	var spine strings.Builder

	// Add chapters to manifest and spine
	for _, chapter := range w.chapters {
		manifest.WriteString(fmt.Sprintf(`    <item id="%s" href="%s" media-type="application/xhtml+xml"/>`,
			chapter.ID, chapter.Filename))
		manifest.WriteString("\n")

		spine.WriteString(fmt.Sprintf(`    <itemref idref="%s"/>`, chapter.ID))
		spine.WriteString("\n")
	}

	customMeta := w.config.GetCustomMetadataElements()

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="%s" unique-identifier="BookId">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:identifier id="BookId">%s</dc:identifier>
    <dc:language>%s</dc:language>
    <dc:date>%s</dc:date>
%s%s  </metadata>
  <manifest>
%s    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="css" href="styles.css" media-type="text/css"/>
  </manifest>
  <spine toc="ncx">
%s  </spine>
</package>`, w.config.Version, w.title, w.config.Creator, w.uuid, w.config.Language,
		time.Now().Format("2006-01-02"),
		w.config.GetMetadataElement("subject", w.config.Subject),
		customMeta, manifest.String(), spine.String())
}

// getTocNCX returns the toc.ncx content
func (w *EPUBWriter) getTocNCX() string {
	var navPoints strings.Builder

	for i, chapter := range w.chapters {
		navPoints.WriteString(fmt.Sprintf(`    <navPoint id="%s" playOrder="%d">
      <navLabel>
        <text>%s</text>
      </navLabel>
      <content src="%s"/>
    </navPoint>`, chapter.ID, i+1, escapeXML(chapter.Title), chapter.Filename))
		navPoints.WriteString("\n")
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="%s"/>
    <meta name="dtb:depth" content="1"/>
    <meta name="dtb:totalPageCount" content="0"/>
    <meta name="dtb:maxPageNumber" content="0"/>
  </head>
  <docTitle>
    <text>%s</text>
  </docTitle>
  <navMap>
%s  </navMap>
</ncx>`, w.uuid, escapeXML(w.title), navPoints.String())
}

// generateUUID returns a unique identifier string for the EPUB using the current Unix nanosecond timestamp.
func generateUUID() string {
	return fmt.Sprintf("urn:uuid:%d", time.Now().UnixNano())
}

// escapeXML returns a copy of the input string with XML special characters escaped.
func escapeXML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&apos;")
	return text
}

// Result returns the final EPUB content (not applicable for ZIP-based format)
func (w *EPUBWriter) Result() string {
	return "" // EPUB is written directly to the ZIP
}

// GetLastError returns the last error that occurred during EPUB generation
func (w *EPUBWriter) GetLastError() error {
	return w.lastError
}

// Reset clears the writer state for reuse
func (w *EPUBWriter) Reset() {
	w.htmlWriter.Reset()
	w.chapters = make([]Chapter, 0)
	w.currentChapter = nil
	w.title = ""
	w.uuid = generateUUID()
	w.lastError = nil
}

// SetOutput sets the output destination
func (w *EPUBWriter) SetOutput(writer io.Writer) {
	w.output = writer
	w.zipWriter = zip.NewWriter(writer)
}

// EPUBWriterV2 implements the WriterV2 interface for EPUB output
type EPUBWriterV2 struct {
	stats      WriterStats
	buffer     *strings.Builder
	epubWriter *EPUBWriter
	binaryData []byte // Store binary EPUB data instead of string
}

// Handle processes a single event with error handling
func (w *EPUBWriterV2) Handle(event streaming.Event) error {
	if w.epubWriter == nil {
		// Initialize with buffer if not set
		w.buffer = &strings.Builder{}
		w.epubWriter = NewEPUBWriter(w.buffer)
	}

	w.epubWriter.Handle(event)
	w.stats.EventsProcessed++

	// Check for errors in the underlying writer
	if err := w.epubWriter.GetLastError(); err != nil {
		return fmt.Errorf("EPUB generation error: %w", err)
	}

	switch event.Kind {
	case streaming.Text:
		w.stats.TextChars += len(event.TextContent)
	case streaming.Image:
		w.stats.Images++
	case streaming.StartTable:
		w.stats.Tables++
	case streaming.StartHeading:
		w.stats.Headings++
	case streaming.StartList:
		w.stats.Lists++
	}

	return nil
}

// Flush finalizes any pending operations and returns the result
func (w *EPUBWriterV2) Flush() (string, error) {
	if w.buffer == nil {
		return "", nil
	}

	// Check for any final errors
	if err := w.epubWriter.GetLastError(); err != nil {
		return "", fmt.Errorf("EPUB generation error: %w", err)
	}

	// Convert binary data to string safely using base64 encoding for transport
	// Note: This is still not ideal - the interface should be changed to handle binary data
	w.binaryData = []byte(w.buffer.String())
	return w.buffer.String(), nil // For now, return as-is to maintain interface compatibility
}

// Reset clears the writer state for reuse
func (w *EPUBWriterV2) Reset() {
	w.buffer = &strings.Builder{}
	w.epubWriter = NewEPUBWriter(w.buffer)
	w.stats = WriterStats{}
	w.binaryData = nil
}

// Stats returns processing statistics
func (w *EPUBWriterV2) Stats() WriterStats {
	return w.stats
}
