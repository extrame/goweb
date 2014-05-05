package goweb

import (
	"encoding/json"
	"errors"
)

// Interface describing an object responsible for
// handling transformed/formatted response data
type Formatter interface {
	// method to transform response
	Format(context *Context, input interface{}) ([]uint8, error)
	// method that decides if this formatter will be used
	Match(*Context) bool
	// init the formatter
	Init()
}

// Internal collection of formatters
var formatters []Formatter = make([]Formatter, 0)
var default_formatter Formatter

// Adds a formatter decider
func AddFormatter(fs ...Formatter) {
	for _, f := range fs {
		f.Init()
	}
	formatters = append(formatters, fs...)
}

// Clears all formatters (including default internal ones)
func ClearFormatters() {
	formatters = nil
}

// Gets the relevant formatter for the specified context or
// returns an error if no formatter is found
func GetFormatter(cx *Context) (Formatter, error) {

	// warn if someone cleared them all out
	if formatters == nil {
		return nil, errors.New("There are no formatters set")
	}

	// check each formatter for a match
	for _, formatter := range formatters {
		if formatter.Match(cx) {
			return formatter, nil
		}
	}

	if default_formatter != nil {
		return default_formatter, nil
	}

	// none found
	return nil, errors.New("No suitable Formatter could be found to deal with that request, consider calling ConfigureDefaultFormatters() or AddFormatter().  See http://code.google.com/p/goweb/wiki/APIDocumentation#Formatters")

}

/*

	Default internal formatters

*/

// Formatter for JSON
type JsonFormatter struct{}

// Readies response and converts input data into JSON
func (f *JsonFormatter) Format(cx *Context, input interface{}) ([]uint8, error) {
	// marshal json
	output, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	// if callback param is given format as JSONP
	if callback := cx.GetCallback(); callback != "" {
		// get the context var
		requestContext := cx.GetRequestContext()

		// wrap in js function
		outputString := callback + "(" + string(output)

		// pass the request context as the second param
		if requestContext != "" {
			outputString = outputString + ", \"" + requestContext + "\")"
		} else {
			outputString = outputString + ")"
		}
		// set the new content type
		cx.ResponseWriter.Header().Set("Content-Type", JSONP_CONTENT_TYPE)

		// convert back
		output = []uint8(outputString)

	} else {
		// normal json content type
		cx.ResponseWriter.Header().Set("Content-Type", "text/plain")
		cx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		cx.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Authorization")
	}

	return output, nil
}

// Gets the "application/json" content type
func (f *JsonFormatter) Match(cx *Context) bool {
	return cx.Format == JSON_FORMAT
}

func (f *JsonFormatter) Init() {
}

// Formatter for JSON
type RawFormatter struct{}

// Readies response and converts input data into JSON
func (f *RawFormatter) Format(cx *Context, input interface{}) (output []uint8, err error) {

	switch t := input.(type) {
	case string:
		output = []byte(t)
	case []byte:
		output = t
	default:
		output, err = json.Marshal(input)
		if err != nil {
			return nil, err
		}
	}

	cx.ResponseWriter.Header().Set("Content-Type", "text/html")

	return output, nil
}

// Gets the "application/json" content type
func (f *RawFormatter) Match(cx *Context) bool {
	return cx.Format == RAW_FORMAT
}

func (f *RawFormatter) Init() {
}

// Adds the default formatters to goweb so that
func ConfigureDefaultFormatters() {
	AddFormatter(new(JsonFormatter))
}
