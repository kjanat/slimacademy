package events

import (
	"strings"

	"github.com/kjanat/slimacademy/internal/models"
)

// convertTextStyle converts a Google Docs TextStyle to our bit-flag Style
func convertTextStyle(textStyle models.TextStyle) (Style, string) {
	var style Style
	var linkURL string

	if textStyle.Bold != nil && *textStyle.Bold {
		style |= Bold
	}
	if textStyle.Italic != nil && *textStyle.Italic {
		style |= Italic
	}
	if textStyle.Underline != nil && *textStyle.Underline {
		style |= Underline
	}
	if textStyle.Strikethrough != nil && *textStyle.Strikethrough {
		style |= Strike
	}

	// Handle subscript/superscript (mutually exclusive - prefer superscript)
	if isSuperscript(textStyle) {
		style |= Sup
	} else if isSubscript(textStyle) {
		style |= Sub
	}

	// Handle highlight (checking for background color)
	if hasHighlight(textStyle) {
		style |= Highlight
	}

	// Handle links
	if textStyle.Link != nil && textStyle.Link.URL != "" {
		style |= Link
		linkURL = textStyle.Link.URL
	}

	return style, linkURL
}

// isSubscript checks if the text style indicates subscript formatting
func isSubscript(style models.TextStyle) bool {
	if style.BaselineOffset != nil {
		offset := strings.ToLower(*style.BaselineOffset)
		return strings.Contains(offset, "sub") || strings.HasPrefix(*style.BaselineOffset, "-")
	}
	return false
}

// isSuperscript checks if the text style indicates superscript formatting
func isSuperscript(style models.TextStyle) bool {
	if style.BaselineOffset != nil {
		offset := strings.ToLower(*style.BaselineOffset)
		return strings.Contains(offset, "super") ||
			(strings.HasPrefix(*style.BaselineOffset, "+") ||
				(!strings.HasPrefix(*style.BaselineOffset, "-") && *style.BaselineOffset != "" && *style.BaselineOffset != "0"))
	}
	return false
}

// hasHighlight checks if the text style has background color (highlight)
func hasHighlight(style models.TextStyle) bool {
	// Google Docs uses backgroundColor for highlighting
	return style.BackgroundColor != nil
}
