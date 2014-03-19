package goweb

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Fill a struct `v` from the values in `form`
func UnmarshalForm(form *map[string][]string, v interface{}, autofill bool) error {
	// check v is valid
	rv := reflect.ValueOf(v).Elem()
	// dereference pointer
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {

		// for each struct field on v
		unmarshalStructInForm("", form, rv, 0, autofill)
	} else if rv.Kind() == reflect.Map && !rv.IsNil() {
		// for each form value add it to the map
		for k, v := range *form {
			if len(v) > 0 {
				rv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
			}
		}
	} else {
		return fmt.Errorf("v must point to a struct or a non-nil map type")
	}
	return nil
}

func unmarshalStructInForm(context string, form *map[string][]string, rvalue reflect.Value, offset int, autofill bool) (err error) {

	if rvalue.Type().Kind() == reflect.Ptr {
		rvalue = rvalue.Elem()
	}
	rtype := rvalue.Type()

	success := false

	for i := 0; i < rtype.NumField(); i++ {
		id, form_values, tag, increaseOffset := getFormField(context, form, rtype.Field(i), offset)
		var used_offset = 0
		if increaseOffset {
			used_offset = offset
		}
		switch rtype.Field(i).Type.Kind() {
		case reflect.Struct:
			if rtype.Field(i).Type.PkgPath() == "time" && rtype.Field(i).Type.Name() == "Time" {
				if len(tag) > 0 && tag[0] == "fillby(now)" && autofill {
					rvalue.Field(i).Set(reflect.ValueOf(time.Now()))
				} else if len(form_values) > 0 {
					time, err := time.Parse(time.RFC3339, form_values[used_offset])
					if err == nil {
						rvalue.Field(i).Set(reflect.ValueOf(time))
					} else {
						fmt.Println(err)
						return err
					}
				}
			} else {
				unmarshalStructInForm(id, form, rvalue.Field(i), 0, autofill)
			}
		case reflect.Slice:

			subRType := rtype.Field(i).Type.Elem()
			switch subRType.Kind() {
			case reflect.Struct:
				rvalueTemp := reflect.MakeSlice(rtype.Field(i).Type, 0, 0)
				subRValue := reflect.New(subRType)
				offset := 0
				for {
					err = unmarshalStructInForm(id, form, subRValue, offset, autofill)
					if err != nil {
						break
					}

					offset++
					rvalueTemp = reflect.Append(rvalueTemp, subRValue.Elem())
				}
				rvalue.Field(i).Set(rvalueTemp)
			default:
				len_fv := len(form_values)
				rvalue = reflect.MakeSlice(rtype.Field(i).Type, len_fv, len_fv)
				for i := 0; i < len_fv; i++ {
					unmarshalField(context, form, rvalue.Field(i), form_values[i], autofill, tag)
				}
			}
		default:
			if len(form_values) > 0 && used_offset < len(form_values) {
				unmarshalField(context, form, rvalue.Field(i), form_values[used_offset], autofill, tag)
				success = true
			}
		}
	}
	if !success {
		return errors.New("no more element")
	}
	return
}

func getFormField(prefix string, form *map[string][]string, t reflect.StructField, offset int) (string, []string, []string, bool) {
	tags := strings.Split(t.Tag.Get("form"), ",")
	tag := tags[0]
	var values = (*form)[tag]
	var increaseOffset = true
	if len(tags) == 0 || tags[0] == "" {
		tag = t.Name
		values = (*form)[tag]
	}
	if prefix != "" {
		tag1 := prefix + "[" + tag + "]"
		tag2 := fmt.Sprintf(prefix+"[%d]"+"["+tag+"]", offset)

		values = (*form)[tag1]
		tag = tag1

		if (*form)[tag2] != nil {
			values = (*form)[tag2]
			tag = tag2
			increaseOffset = false
		}
	}
	return tag, values, tags[1:], increaseOffset
}

func unmarshalField(context string, form *map[string][]string, v reflect.Value, form_value string, autofill bool, tags []string) error {
	// string -> type conversion
	switch v.Kind() {
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		// convert to Int
		// convert to Int64
		if i, err := strconv.ParseInt(form_value, 10, 64); err == nil {
			v.SetInt(i)
		}
	case reflect.String:
		// copy string
		v.SetString(form_value)
	case reflect.Float64:
		if f, err := strconv.ParseFloat(form_value, 64); err == nil {
			v.SetFloat(f)
		}
	case reflect.Float32:
		if f, err := strconv.ParseFloat(form_value, 32); err == nil {
			v.SetFloat(f)
		}
	case reflect.Bool:
		// the following strings convert to true
		// 1,true,on,yes
		fv := form_value
		if fv == "1" || fv == "true" || fv == "on" || fv == "yes" {
			v.SetBool(true)
		}
	default:
		fmt.Println("unknown type", v.Kind())
	}
	return nil
}
