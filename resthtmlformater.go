package goweb

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

var document_root string

// Formatter for JSON
type RestHtmlFormattor struct{}

type RestModelTemplate struct {
	template.Template
}

var rest_model = template.New("REST_HTTP_ROOT")

func parseFileWithName(parent *template.Template,name string, filepath string) error {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	s := string(b)
	// First template becomes return value if not already defined,
	// and we use that one for subsequent New calls to associate
	// all the templates together. Also, if this file has the same name
	// as t, this file becomes the contents of t, so
	//  t, err := New(name).Funcs(xxx).ParseFiles(name)
	// works. Otherwise we create a new template associated with t.
	var tmpl *template.Template
	if name == parent.Name() || name == "" {
		tmpl = parent
	} else {
		tmpl = parent.New(name)
	}
	_, err = tmpl.Parse(s)
	if err != nil {
		return err
	}
	return nil
}

func initModelTemplate(url string) *template.Template{
	temp := rest_model.New(url)
	//scan for the helpers
	filepath.Walk(filepath.Join(document_root, url, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) {
			fmt.Println("Parse helper:", path)
			e := parseFileWithName(temp,filepath.Join("model", info.Name()), path)
			if e != nil {
				fmt.Printf("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
	return temp
}

func initGlobalTemplate() {
	//scan for the helpers
	filepath.Walk(filepath.Join(document_root, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) {
			fmt.Println("Parse helper:", path)
			e := parseFileWithName(rest_model,filepath.Join("global", info.Name()), path)
			if e != nil {
				fmt.Printf("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
}

//Set the root for the rest html formating, formating is based on the method name(lower case)
func SetDocumentRoot(root string) {
	document_root = root
	initGlobalTemplate()
}

func getRestModelByContext(cx *Context) *template.Template{
	t := rest_model.Lookup(cx.Rest.Url)
	
	if t == nil {
		return initModelTemplate(cx.Rest.Url)
	}
	
	return t
}

func getMethodTemplate(m *template.Template,rest *RestContext) *template.Template {
	t := m.Lookup(rest.Method+".html")
	var err error
	if t == nil {
		t, err = m.New(rest.Method + ".html").ParseFiles(filepath.Join(document_root, rest.Url, rest.Method+".html"))
		if err != nil {
			fmt.Println("ERROR template.ParseFile: %v", err)
		}
	}
	return t
}

// Readies response and converts input data into JSON
func (f *RestHtmlFormattor) Format(cx *Context, input interface{}) ([]uint8, error) {
	//get the document root dir
	model := getRestModelByContext(cx)
	
	temp := getMethodTemplate(model,&cx.Rest)
	
	var err error

	buffer := new(bytes.Buffer)
	std_input, ok := input.(*standardResponse)
	if ok {
		err = temp.Execute(buffer, std_input.D)
	} else {
		err = temp.Execute(buffer, input)
	}
	if err != nil {
		fmt.Printf("ERROR template.Execute: %v", err)
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
