package goweb

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"
)

// types that impliment RequestDecoder can unmarshal
// the request body into an apropriate type/struct
type RequestDecoder interface {
	Unmarshal(cx *Context, v interface{}, autofill bool) error
}

// a JSON decoder for request body (just a wrapper to json.Unmarshal)
type JsonRequestDecoder struct{}

func (d *JsonRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) (err error) {
	// read body
	if cx.RequestData == nil {
		cx.RequestData, err = ioutil.ReadAll(cx.Request.Body)
		if err != nil {
			return err
		}
	}
	return json.Unmarshal(cx.RequestData, v)
}

// an XML decoder for request body
type XmlRequestDecoder struct{}

func (d *XmlRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	// read body
	data, err := ioutil.ReadAll(cx.Request.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// a form-enc decoder for request body
type FormRequestDecoder struct{}

func (d *FormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	if cx.Request.Form == nil {
		cx.Request.ParseForm()
	}
	var values *map[string][]string
	values = (*map[string][]string)(&cx.Request.Form)
	return UnmarshalForm(values, v, autofill)
}

// a form-enc decoder for request body
type MultiFormRequestDecoder struct{}

func (d *MultiFormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	cx.Request.ParseMultipartForm(32 << 20)
	values := (map[string][]string)(cx.Request.Form)
	for k, v := range cx.Request.MultipartForm.Value {
		values[k] = v
	}
	return UnmarshalForm(&values, v, autofill)
}

// map of Content-Type -> RequestDecoders
var decoders map[string]RequestDecoder = map[string]RequestDecoder{
	"application/json":                  new(JsonRequestDecoder),
	"application/xml":                   new(XmlRequestDecoder),
	"text/xml":                          new(XmlRequestDecoder),
	"application/x-www-form-urlencoded": new(FormRequestDecoder),
	"text/plain":                        new(FormRequestDecoder),
	"multipart/form-data":               new(MultiFormRequestDecoder),
}

// goweb.Context Helper function to fill a variable with the contents
// of the request body. The body will be decoded based
// on the content-type and an apropriate RequestDecoder
// automatically selected
func (cx *Context) Fill(v interface{}, autofill bool) error {
	// get content type
	ct := cx.Request.Header.Get("Content-Type")
	// default to urlencoded
	if ct == "" {
		ct = "application/x-www-form-urlencoded"
	}
	return cx.FillAs(v, autofill, ct)
}

func (cx *Context) FillAs(v interface{}, autofill bool, ct string) error {
	// ignore charset (after ';')
	ct = strings.Split(ct, ";")[0]
	// get request decoder
	decoder, ok := decoders[ct]
	if ok != true {
		return fmt.Errorf("Cannot decode request for %s data", ct)
	}
	// decode
	err := decoder.Unmarshal(cx, v, autofill)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// all clear
	return nil
}
