package goweb

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

var document_root string

// Formatter for JSON
type RestHtmlFormattor struct{}

var rest_template map[string]*template.Template = map[string]*template.Template{}

func getPathByContext(cx *Context) (string, string) {
	var file_path string
	id, ok := cx.PathParams["id"]
	if ok {
		file_path = strings.Replace(cx.Request.URL.Path, strings.ToLower(id+"."+cx.Format), strings.ToLower(cx.RestMethod+"."+cx.Format), -1)
	} else {
		if cx.RestMethod == NEW_REST_METHOD {
			file_path = cx.Request.URL.Path
		} else {
			file_path = strings.Replace(cx.Request.URL.Path, strings.ToLower("."+cx.Format), strings.ToLower("/"+cx.RestMethod+"."+cx.Format), -1)
		}
	}
	res := filepath.Join(document_root, file_path)
	return res, file_path
}

//Set the root for the rest html formating, formating is based on the method name(lower case)
func SetDocumentRoot(root string) {
	document_root = root
}

// Readies response and converts input data into JSON
func (f *RestHtmlFormattor) Format(cx *Context, input interface{}) ([]uint8, error) {
	//get the document root dir
	path, name := getPathByContext(cx)
	var err error

	cur_template := rest_template[name]

	if cur_template == nil {
		fmt.Println("init template")
		cur_template, err = template.ParseFiles(path)
		rest_template[name] = cur_template
	}

	if err != nil {
		fmt.Println("ERROR template.ParseFile: %v", err)
		return nil, err
	}

	buffer := new(bytes.Buffer)
	std_input, ok := input.(*standardResponse)
	if ok {
		err = cur_template.Execute(buffer, std_input.D)
	} else {
		err = cur_template.Execute(buffer, input)
	}
	if err != nil {
		fmt.Println("ERROR template.Execute: %v", err)
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Gets the "application/json" content type
func (f *RestHtmlFormattor) Match(cx *Context) bool {
	return cx.Format == HTML_FORMAT
}
