package goweb

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Wraps a controllerFunc to catch any panics, log them and
// respond with an appropriate error
func safeControllerFunc(controllerFunc func(*Context)) func(*Context) {
	return func(cx *Context) {
		defer func() {
			if err := recover(); err != nil {
				lines := strings.Split(string(debug.Stack()), "\n")[4:]
				log.Print("panic: ", err, "\n", strings.Join(lines, "\n"))
				cx.RespondWithErrorCode(500)
			}
		}()
		controllerFunc(cx)
	}
}

// Maps a new route to a controller (with optional RouteMatcherFuncs)
// and returns the new route
func Map(path string, controller Controller, matcherFuncs ...RouteMatcherFunc) *Route {
	return DefaultRouteManager.Map(path, controller, matcherFuncs...)
}

// Maps a new route to a function (with optional RouteMarcherFuncs)
// and returns the new route
func MapFunc(path string, controllerFunc func(*Context), matcherFuncs ...RouteMatcherFunc) *Route {
	return DefaultRouteManager.MapFunc(path, safeControllerFunc(controllerFunc), matcherFuncs...)
}

// Maps an entire RESTful set of routes to the specified RestController
// You only have to specify the methods that you require see rest_controller.go
// for the list of interfaces that can be satisfied
func MapRest(pathPrefix string, controller RestController) {
	var rest RestContext
	rest.Url = pathPrefix

	var pathPrefixWithId string = pathPrefix + "/{id}"

	// OPTIONS /resource
	if rc, ok := controller.(RestOptions); ok {
		MapFunc(pathPrefix, func(c *Context) {
			c.Rest = rest
			c.Rest.Method = OPTIONS_REST_METHOD
			rc.Options(c)
		}, OptionsMethod)
	}
	// GET /resource/new
	if rc, ok := controller.(RestNewer); ok {
		MapFunc(pathPrefix+"/new", func(c *Context) {
			c.Rest = rest
			c.Rest.Method = NEW_REST_METHOD
			rc.New(c)
		}, GetMethod)
	}
	// GET /resource/{id};edit
	if rc, ok := controller.(RestEditor); ok {
		MapFunc(pathPrefixWithId+"/edit", func(c *Context) {
			c.Rest = rest
			c.Rest.Method = EDIT_REST_METHOD
			rc.Edit(c.PathParams["id"], c)
		}, GetMethod)
	}
	// GET /resource/{id}
	if rc, ok := controller.(RestReader); ok {
		MapFunc(pathPrefixWithId, func(c *Context) {
			c.Rest = rest
			c.Rest.Method = READ_REST_METHOD
			rc.Read(c.PathParams["id"], c)
		}, GetMethod)
	}

	// GET /resource
	if rc, ok := controller.(RestManyReader); ok {
		MapFunc(pathPrefix, func(c *Context) {
			c.Rest = rest
			c.Rest.Method = READMANY_REST_METHOD
			rc.ReadMany(c)
		}, GetMethod)
	}
	// PUT /resource/{id}
	if rc, ok := controller.(RestUpdater); ok {
		MapFunc(pathPrefixWithId, func(c *Context) {
			c.Rest = rest
			c.Rest.Method = UPDATE_REST_METHOD
			rc.Update(c.PathParams["id"], c)
		}, PutMethod)
	}
	// PUT /resource
	if rc, ok := controller.(RestManyUpdater); ok {
		MapFunc(pathPrefix, func(c *Context) {
			c.Rest = rest
			c.Rest.Method = UPDATEMANY_REST_METHOD
			rc.UpdateMany(c)
		}, PutMethod)
	}
	// DELETE /resource/{id}
	if rc, ok := controller.(RestDeleter); ok {
		MapFunc(pathPrefixWithId, func(c *Context) {
			c.Rest = rest
			c.Rest.Method = DELETE_REST_METHOD
			rc.Delete(c.PathParams["id"], c)
		}, DeleteMethod)
	}
	// DELETE /resource
	if rc, ok := controller.(RestManyDeleter); ok {
		MapFunc(pathPrefix, func(c *Context) {
			c.Rest = rest
			c.Rest.Method = DELETEMANY_REST_METHOD
			rc.DeleteMany(c)
		}, DeleteMethod)
	}
	// CREATE /resource
	if rc, ok := controller.(RestCreator); ok {
		MapFunc(pathPrefix, func(c *Context) {
			c.Rest = rest
			c.Rest.Method = CREATE_REST_METHOD
			rc.Create(c)
		}, PostMethod)
	}
}

// Maps a path to a static directory
func MapStatic(pathPrefix string, rootDirectory string) {
	MapFunc(pathPrefix, func(cx *Context) {
		path := cx.Request.URL.Path
		path = filepath.Join(rootDirectory, path)
		fmt.Println("static:", path)
		http.ServeFile(cx.ResponseWriter, cx.Request, path)
	})
}

func MapFormattedStatic(pathPrefix string, obj interface{}) {
	MapFunc(pathPrefix, func(cx *Context) {
		cx.WriteResponse(obj, 200)
	})
}
