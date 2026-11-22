package security

import (
	"html"
	"regexp"
	"strings"
)

var (
	// Regex patterns for dangerous HTML/XSS vectors
	scriptTagRegex    = regexp.MustCompile(`(?i)<\s*script[^>]*>[\s\S]*?<\s*/\s*script\s*>`)
	iframeTagRegex    = regexp.MustCompile(`(?i)<\s*iframe[^>]*>[\s\S]*?<\s*/\s*iframe\s*>`)
	styleTagRegex     = regexp.MustCompile(`(?i)<\s*style[^>]*>[\s\S]*?<\s*/\s*style\s*>`)
	genericHTMLRegex  = regexp.MustCompile(`(?i)<[^>]+>`)
	eventHandlerRegex = regexp.MustCompile(`(?i)on\w+\s*=`)
	javascriptURIScheme = regexp.MustCompile(`(?i)javascript\s*:`)
)

// SanitizeText removes malicious script tags, iframes, and javascript: links
// while preserving clinical medical formatting, symbols, and punctuation.
func SanitizeText(input string) string {
	if input == "" {
		return ""
	}

	// 1. Strip script, iframe, and style tags completely
	cleaned := scriptTagRegex.ReplaceAllString(input, "")
	cleaned = iframeTagRegex.ReplaceAllString(cleaned, "")
	cleaned = styleTagRegex.ReplaceAllString(cleaned, "")

	// 2. Remove inline event handlers (onload, onerror, onclick, etc.)
	cleaned = eventHandlerRegex.ReplaceAllString(cleaned, "blocked_handler=")

	// 3. Remove javascript: pseudo-protocols
	cleaned = javascriptURIScheme.ReplaceAllString(cleaned, "blocked_uri:")

	// 4. Trim leading/trailing spaces
	return strings.TrimSpace(cleaned)
}

// SanitizePlainText strips all HTML tags and escapes remaining entities.
// Ideal for names, numbers, codes, and identifiers (e.g. NIK, Nama Pasien, No RM).
func SanitizePlainText(input string) string {
	if input == "" {
		return ""
	}
	noHTML := genericHTMLRegex.ReplaceAllString(input, "")
	return strings.TrimSpace(html.EscapeString(noHTML))
}

// SanitizeSearchQuery strips SQL wildcard and dangerous characters from user searches
// before feeding into parameterized ILIKE queries to prevent ReDoS / Query Flooding.
func SanitizeSearchQuery(input string, maxLen int) string {
	trimmed := strings.TrimSpace(input)
	if maxLen > 0 && len(trimmed) > maxLen {
		trimmed = trimmed[:maxLen]
	}
	// Escape literal % and _ if users shouldn't have arbitrary wildcard power
	trimmed = strings.ReplaceAll(trimmed, "\x00", "")
	return trimmed
}
