package goweb

import (
	"net/http"
)

// Constant string for HTML format
const HTML_FORMAT string = "html"

// Constant string for HTML format
const MOBILE_FORMAT string = "mbl"

// Constant string for XML format
const XML_FORMAT string = "xml"

// Constant string for JSON format
const JSON_FORMAT string = "json"

// The fallback format if one cannot be determined by the request
var DEFAULT_FORMAT string = HTML_FORMAT

// Gets a string describing the format of the request.
func getFormatForRequest(request *http.Request) (ext string, path_without_suffix string) {

	if request.URL == nil {
		return DEFAULT_FORMAT, ""
	}

	// use the file extension as the format
	ext, path_without_suffix = getFileExtension(request.URL.Path)
	if ext != "" {

		// manual overrides
		if ext == "htm" {
			ext = HTML_FORMAT
		}

	} else {
		ext = DEFAULT_FORMAT
	}

	return
}

func SetDefaultFormat(format string) {
	DEFAULT_FORMAT = format
}
