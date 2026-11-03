package attachments

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const textPlain = "text/plain"

var binaryTypes = map[string]bool{
	"image/png":       true,
	"image/jpeg":      true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
}

var textTypes = map[string]string{
	".md":       "text/markdown",
	".markdown": "text/markdown",
	".json":     "application/json",
	".csv":      "text/csv",
	".xml":      "application/xml",
	".yaml":     "application/yaml",
	".yml":      "application/yaml",
	".html":     "text/html",
	".htm":      "text/html",
}

var refused = map[string]bool{".svg": true, ".svgz": true}

// Detect names the media type of an upload from its bytes and its name. The
// browser's own label is never trusted: it calls a .ts file video/mp2t and
// sends nothing at all for a Dockerfile. Images and PDFs are recognised by
// their signature, anything else must be valid UTF-8 text, and SVG is refused
// because it is a document that can run script.
func Detect(name string, data []byte) (string, bool) {
	ext := strings.ToLower(filepath.Ext(name))
	if refused[ext] {
		return "", false
	}
	sniffed, _, _ := strings.Cut(http.DetectContentType(data), ";")
	if binaryTypes[sniffed] {
		return sniffed, true
	}
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return "", false
	}
	if known, ok := textTypes[ext]; ok {
		return known, true
	}
	return textPlain, true
}

// served is the Content-Type a stored file is answered with. Text of any kind
// goes out as plain text, so an uploaded page never renders on Bulle's origin.
func served(mediaType string) string {
	if binaryTypes[mediaType] {
		return mediaType
	}
	return "text/plain; charset=utf-8"
}
