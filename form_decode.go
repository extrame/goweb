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
		unmarshalStructInForm("", form, rv, 0, autofill, false)
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

func unmarshalStructInForm(context string, form *map[string][]string, rvalue reflect.Value, offset int, autofill bool, inarray bool) (err error) {

	if rvalue.Type().Kind() == reflect.Ptr {

		rvalue = rvalue.Elem()
	}
	rtype := rvalue.Type()

	success := false

	for i := 0; i < rtype.NumField(); i++ {
		id, form_values, tag, increaseOffset := getFormField(context, form, rtype.Field(i), offset, inarray)
		var used_offset = 0
		if increaseOffset {
			used_offset = offset
		}
		switch rtype.Field(i).Type.Kind() {
		case reflect.Ptr: //TODO if the ptr point to a basic data, it will crash
			val := rvalue.Field(i)
			typ := rtype.Field(i).Type.Elem()
			if val.IsNil() {
				val.Set(reflect.New(typ))
			}
			if err := fill_struct(typ, form, rvalue.Field(i), id, form_values, tag, used_offset, autofill); err != nil {
				fmt.Println(err)
				return err
			} else {
				break
			}
		case reflect.Struct:
			if err := fill_struct(rtype.Field(i).Type, form, rvalue.Field(i), id, form_values, tag, used_offset, autofill); err != nil {
				fmt.Println(err)
				return err
			} else {
				break
			}
		case reflect.Slice:
			subRType := rtype.Field(i).Type.Elem()
			switch subRType.Kind() {
			case reflect.Struct:
				rvalueTemp := reflect.MakeSlice(rtype.Field(i).Type, 0, 0)
				subRValue := reflect.New(subRType)
				offset := 0
				for {
					err = unmarshalStructInForm(id, form, subRValue, offset, autofill, true)
					if err != nil {
						fmt.Println(err)
						break
					}

					offset++
					rvalueTemp = reflect.Append(rvalueTemp, subRValue.Elem())
				}
				rvalue.Field(i).Set(rvalueTemp)
			default:
				len_fv := len(form_values)
				rvnew := reflect.MakeSlice(rtype.Field(i).Type, len_fv, len_fv)
				for j := 0; j < len_fv; j++ {
					unmarshalField(context, form, rvnew.Index(j), form_values[i], autofill, tag)
				}
				rvalue.Field(i).Set(rvnew)
			}
		default:
			if len(form_values) > 0 && used_offset < len(form_values) {
				unmarshalField(context, form, rvalue.Field(i), form_values[used_offset], autofill, tag)
				success = true
			}
		}
	}
	fmt.Println(rvalue.Interface(), success)
	if !success {
		return errors.New("no more element")
	}
	return nil
}

func getFormField(prefix string, form *map[string][]string, t reflect.StructField, offset int, inarray bool) (string, []string, []string, bool) {
	tags := strings.Split(t.Tag.Get("form"), ",")
	tag := tags[0]
	var values = (*form)[tag]
	var increaseOffset = true
	if len(tags) == 0 || tags[0] == "" {
		tag = t.Name
		values = (*form)[tag]
	}
	if prefix != "" {
		if inarray {
			increaseOffset = false
			tag = fmt.Sprintf(prefix+"[%d]"+"["+tag+"]", offset)
		} else {
			increaseOffset = true
			tag = prefix + "[" + tag + "]"
		}
		values = (*form)[tag]
		fmt.Println(tag, values)
	}
	return tag, values, tags[1:], increaseOffset
}

func fill_struct(typ reflect.Type, form *map[string][]string, val reflect.Value, id string, form_values []string, tag []string, used_offset int, autofill bool) error {
	if typ.PkgPath() == "time" && typ.Name() == "Time" {
		if len(tag) > 0 && tag[0] == "fillby(now)" && autofill {
			val.Set(reflect.ValueOf(time.Now()))
		} else if len(form_values) > 0 {
			time, err := time.Parse(time.RFC3339, form_values[used_offset])
			if err == nil {
				val.Set(reflect.ValueOf(time))
			} else {
				fmt.Println(err)
				return err
			}
		}
	} else {
		unmarshalStructInForm(id, form, val, 0, autofill, false)
	}
	return nil
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
