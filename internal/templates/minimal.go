// Package templates provides minimal HTML templates for SlimAcademy document rendering.
// This package replaces the complex embedded CSS approach with clean, minimal templates
// focusing on readability, accessibility, and performance.
package templates

import (
	"html/template"
	"strings"
)

// MinimalTemplate provides a clean, minimal HTML template system
type MinimalTemplate struct {
	htmlTemplate *template.Template
}

// NewMinimalTemplate creates a new minimal template instance
func NewMinimalTemplate() *MinimalTemplate {
	tmpl := template.Must(template.New("minimal").Parse(minimalHTMLTemplate))
	return &MinimalTemplate{
		htmlTemplate: tmpl,
	}
}

// TemplateData holds data for template rendering
type TemplateData struct {
	Title       string
	Description string
	Content     template.HTML
	Metadata    map[string]string
	HasMetadata bool
	Generator   string
	CSS         template.CSS
}

// DefaultCSS returns minimal, clean CSS styling
func DefaultCSS() template.CSS {
	return template.CSS(minimalCSS)
}

// Render executes the template with provided data
func (mt *MinimalTemplate) Render(data TemplateData) (string, error) {
	// Set defaults
	if data.Generator == "" {
		data.Generator = "Slim Academy"
	}
	if data.CSS == "" {
		data.CSS = DefaultCSS()
	}
	data.HasMetadata = len(data.Metadata) > 0

	var result strings.Builder
	err := mt.htmlTemplate.Execute(&result, data)
	return result.String(), err
}

// minimalHTMLTemplate is a clean, semantic HTML5 template
const minimalHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    {{if .Description}}<meta name="description" content="{{.Description}}">{{end}}
    <meta name="generator" content="{{.Generator}}">
    <style>{{.CSS}}</style>
</head>
<body>
    <main class="document">
        <header class="document-header">
            <h1 class="document-title">{{.Title}}</h1>
            {{if .Description}}<p class="document-description">{{.Description}}</p>{{end}}
            {{if .HasMetadata}}
            <div class="document-metadata">
                {{range $key, $value := .Metadata}}
                <span class="metadata-item">
                    <strong>{{$key}}:</strong> {{$value}}
                </span>
                {{end}}
            </div>
            {{end}}
        </header>

        <div class="document-content">
            {{.Content}}
        </div>
    </main>
</body>
</html>`

// minimalCSS provides clean, readable styling without complexity
const minimalCSS = `/* Reset and base styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    line-height: 1.6;
    color: #333;
    background: #fff;
}

.document {
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem 1rem;
}

/* Header styles */
.document-header {
    border-bottom: 1px solid #e5e5e5;
    padding-bottom: 2rem;
    margin-bottom: 3rem;
}

.document-title {
    font-size: 2.5rem;
    font-weight: 600;
    line-height: 1.2;
    margin-bottom: 0.5rem;
    color: #1a1a1a;
}

.document-description {
    font-size: 1.1rem;
    color: #666;
    margin-bottom: 1rem;
}

.document-metadata {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    font-size: 0.9rem;
    color: #888;
}

.metadata-item {
    padding: 0.25rem 0.5rem;
    background: #f8f9fa;
    border-radius: 4px;
}

/* Content styles */
.document-content {
    font-size: 1rem;
    line-height: 1.7;
}

.document-body {
    display: block;
}

.chapter-section {
    margin-bottom: 3rem;
    padding: 1.5rem 0;
    border-bottom: 1px solid #f0f0f0;
}

.chapter-section:last-child {
    border-bottom: none;
}

.document-content h1,
.document-content h2,
.document-content h3,
.document-content h4,
.document-content h5,
.document-content h6 {
    margin: 2rem 0 1rem 0;
    font-weight: 600;
    line-height: 1.3;
    color: #1a1a1a;
}

.document-content h1 { font-size: 2rem; }
.document-content h2 { font-size: 1.75rem; }
.document-content h3 { font-size: 1.5rem; }
.document-content h4 { font-size: 1.25rem; }
.document-content h5 { font-size: 1.125rem; }
.document-content h6 { font-size: 1rem; }

.document-content p {
    margin: 1rem 0;
}

.document-content ul,
.document-content ol {
    margin: 1rem 0;
    padding-left: 2rem;
}

.document-content li {
    margin: 0.5rem 0;
}

.document-content a {
    color: #0066cc;
    text-decoration: none;
}

.document-content a:hover {
    text-decoration: underline;
}

.document-content strong {
    font-weight: 600;
}

.document-content em {
    font-style: italic;
}

.document-content code {
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    background: #f6f8fa;
    padding: 0.2rem 0.4rem;
    border-radius: 3px;
    font-size: 0.9em;
}

.document-content img {
    max-width: 100%;
    height: auto;
    margin: 1rem 0;
    border-radius: 4px;
}

.document-content table {
    width: 100%;
    border-collapse: collapse;
    margin: 1.5rem 0;
}

.document-content th,
.document-content td {
    padding: 0.75rem;
    text-align: left;
    border-bottom: 1px solid #e5e5e5;
}

.document-content th {
    font-weight: 600;
    background: #f8f9fa;
}

.document-content blockquote {
    margin: 1.5rem 0;
    padding: 1rem 1.5rem;
    border-left: 4px solid #e5e5e5;
    background: #f8f9fa;
    font-style: italic;
}

/* Responsive design */
@media (max-width: 768px) {
    .document {
        padding: 1rem 0.75rem;
    }

    .document-title {
        font-size: 2rem;
    }

    .document-metadata {
        flex-direction: column;
        gap: 0.5rem;
    }
}

/* Print styles */
@media print {
    .document {
        max-width: none;
        padding: 0;
    }

    .document-header {
        border-bottom: 2px solid #000;
    }

    .document-content a {
        color: #000;
        text-decoration: underline;
    }
}`
