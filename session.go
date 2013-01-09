package goweb

import (
	"code.google.com/p/gorilla/sessions"
	"net/http"
	"reflect"
	"encoding/json"
	"errors"
)

var store *sessions.CookieStore

func EnableSession(){
	store = sessions.NewCookieStore([]byte("非常保密"))
}

func MarshallInCookieSession(obj interface{},r *http.Request, w http.ResponseWriter){
	typ := reflect.TypeOf(obj).String()
	session, _ := store.Get(r,typ)
	session.Values[typ],_ = json.Marshal(obj)
	session.Save(r,w)
}

func UnmarshallInCookieSession(obj interface{},r *http.Request)error{
	typ := reflect.TypeOf(obj).String()
	session, _ := store.Get(r,typ)
	res := session.Values[typ]
	if res == nil{
		return errors.New("no such record")
	}
	return json.Unmarshal(res.([]byte),obj)
}


func DeleteFromSession(obj interface{},r *http.Request, w http.ResponseWriter){
	typ := reflect.TypeOf(obj).String()
	session, _ := store.Get(r,typ)
	delete(session.Values,typ)
	session.Save(r,w)
}

