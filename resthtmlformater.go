package goweb

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var document_root string

// Formatter for JSON
type RestHtmlFormattor struct {
	root      *template.Template
	models    map[string]*template.Template
	suffix    string
	dotsuffix string
}

type MobileRestHtmlFormattor struct {
	RestHtmlFormattor
}

type RestModelTemplate struct {
	template.Template
}

func (r *RestHtmlFormattor) Init() {
	r.suffix = "html"
	r.dotsuffix = ".html"
	r.init()
}

func (r *MobileRestHtmlFormattor) Init() {
	r.suffix = "mbl"
	r.dotsuffix = ".mbl"
	r.init()
}

func (r *RestHtmlFormattor) init() {
	r.root = template.New("REST_HTTP_ROOT")
	r.root.Funcs(template.FuncMap{"raw": RawHtml})
	r.models = make(map[string]*template.Template)
	r.initGlobalTemplate()
}

// Gets the "application/html" content type
func (f *RestHtmlFormattor) Match(cx *Context) bool {
	return cx.Format == f.suffix
}

func parseFileWithName(parent *template.Template, name string, filepath string) error {
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

func (f *RestHtmlFormattor) initModelTemplate(url string) *template.Template {
	temp, err := f.root.Clone()
	if err == nil {
		temp = temp.New(url)
		//scan for the helpers
		filepath.Walk(filepath.Join(document_root, url, "helper"), func(path string, info os.FileInfo, err error) error {
			if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), f.dotsuffix) {
				fmt.Println("Parse helper:", path)
				name := strings.TrimSuffix(info.Name(), f.dotsuffix)
				e := parseFileWithName(temp, "model/"+name, path)
				if e != nil {
					fmt.Printf("ERROR template.ParseFile: %v", e)
				}
			}
			return nil
		})
		f.models[url] = temp
		return temp
	}
	fmt.Println("error happened", err)
	return nil
}

func (f *RestHtmlFormattor) initGlobalTemplate() {
	f.root.New("")
	//scan for the helpers
	filepath.Walk(filepath.Join(document_root, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), f.dotsuffix) {
			fmt.Println("Parse helper:", path)
			name := strings.TrimSuffix(info.Name(), f.dotsuffix)
			e := parseFileWithName(f.root, "global/"+name, path)
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
}

func (f *RestHtmlFormattor) getRestModelByContext(cx *Context) *template.Template {
	var t *template.Template
	var path string

	if cx.Rest.Url == "" || cx.Rest.Method == "" {
		path = filepath.Join(document_root, cx.PathWithOutSuffix)
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				path = filepath.Join(document_root, cx.PathWithOutSuffix+"index"+f.dotsuffix)
			}
		} else {
			path = path + f.dotsuffix
		}
		var urlpath = cx.PathWithOutSuffix + f.dotsuffix
		// var path = filepath.Join(document_root, urlpath)
		t = f.models[urlpath]

		if t == nil {
			cloned_rest_model, err := f.root.Clone()

			if err == nil {

				info, err := os.Stat(path)

				if err == nil && info.IsDir() {
					path = filepath.Join(path, "index"+f.dotsuffix)
				}

				err = parseFileWithName(cloned_rest_model, urlpath, path)
				if err == nil {
					t = cloned_rest_model.Lookup(urlpath)
				} else {
					fmt.Println("ERROR template.ParseFile: %v", err)
				}
				f.models[urlpath] = t
			}
		}
	} else {
		t = f.models[cx.Rest.Url]

		if t == nil {
			t = f.initModelTemplate(cx.Rest.Url)
		}

		return f.getMethodTemplate(t, &cx.Rest)
	}

	return t
}

func (f *RestHtmlFormattor) getMethodTemplate(m *template.Template, rest *RestContext) *template.Template {
	t := m.Lookup(rest.Method + f.dotsuffix)
	var err error
	if t == nil {
		t, err = m.New(rest.Method + f.dotsuffix).ParseFiles(filepath.Join(document_root, rest.Url, rest.Method+f.dotsuffix))
		if err != nil {
			fmt.Println("ERROR template.ParseFile: %v", err)
		}
	}
	return t
}

// Readies response and converts input data into JSON
func (f *RestHtmlFormattor) Format(cx *Context, input interface{}) ([]uint8, error) {
	//get the document root dir
	cx.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	temp := f.getRestModelByContext(cx)

	var err error

	buffer := new(bytes.Buffer)

	err = temp.Execute(buffer, input)

	if err != nil {
		fmt.Printf("ERROR template.Execute: %v", err)
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func RawHtml(text string) template.HTML { return template.HTML(text) }
