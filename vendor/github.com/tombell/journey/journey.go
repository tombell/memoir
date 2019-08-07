package journey

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type journeyContextKey int

const (
	contextKeyParamsIdx journeyContextKey = iota
	contextKeyParams
)

func pathSegments(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

// Router routes HTTP requests to matching handlers based on patterns.
type Router struct {
	NotFoundHandler http.Handler
	trees           map[string]*node
}

// NewRouter returns a new Router.
func NewRouter() *Router {
	return &Router{
		NotFoundHandler: http.NotFoundHandler(),
		trees:           make(map[string]*node),
	}
}

// Handle adds an HTTP handler for the specified method and route pattern.
func (r *Router) Handle(method, pattern string, handler http.Handler) {
	// The pattern cannot be empty, and must start with a slash, panic.
	if len(pattern) == 0 || pattern[0] != '/' {
		panic(fmt.Errorf("journey: pattern %q must begin with %q", pattern, "/"))
	}

	method = strings.ToLower(method)

	// Get the nodes of the tree for the given method.
	tree, ok := r.trees[method]
	if !ok {
		tree = &node{}
		r.trees[method] = tree
	}

	// Split the URL pattern into segments.
	parts := pathSegments(pattern)

	var s string
	var params map[string]uint16

	for i, part := range parts {
		s += "/"

		// If the segment is a path parameter.
		if len(part) > 0 && part[0] == ':' {
			tree.insertChild(s, params, nil, nil)

			part = part[1:]
			reSep := strings.IndexByte(part, ':')

			var re *regexp.Regexp

			// If the path parameter has a regexp specified.
			if reSep != -1 {
				if name := part[:reSep]; name != "" {
					if params == nil {
						params = make(map[string]uint16)
					}

					params[name] = uint16(i)
				}

				res := part[reSep+1:]

				// Regexp cannot be empty, panic.
				if res == "" {
					panic(fmt.Errorf("journey: pattern %q has empty regular expression", pattern))
				}

				re = regexp.MustCompile(res)
			} else {
				// Parameter name cannot be empty, panic.
				if part == "" {
					panic(fmt.Errorf("journey: pattern %q has anonymous parameter", pattern))
				}

				if params == nil {
					params = make(map[string]uint16)
				}

				params[part] = uint16(i)
			}

			s += ":"

			if i == len(parts)-1 {
				tree.insertChild(s, params, re, handler)
			} else {
				tree.insertChild(s, params, re, nil)
			}
		} else {
			s += part

			if i == len(parts)-1 {
				tree.insertChild(s, params, nil, handler)
			}
		}
	}
}

// HandleFunc is the http.HandleFunc alternative to http.Handler.
func (r *Router) HandleFunc(method, pattern string, fn http.HandlerFunc) {
	r.Handle(method, pattern, fn)
}

// ServeHTTP routes an incoming http.Request based on method and path.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := strings.ToLower(req.Method)

	if tree, ok := r.trees[method]; ok {
		// Lookup the deepest child node for the request URL.
		child := tree.findChild(req.URL.Path)

		if child != nil && child.handler != nil {
			// Add the param indexes from the child node into the request
			// context. The parameters will be parsed from the path and put into
			// the request context when journey.Param is called for the request.
			if child.params != nil {
				req = req.WithContext(context.WithValue(req.Context(), contextKeyParamsIdx, child.params))
			}

			child.handler.ServeHTTP(w, req)
			return
		}
	}

	// If no handler is found for the method and path, call the not found
	// handler.
	r.NotFoundHandler.ServeHTTP(w, req)
}

// Param returns a path parameter from the request context. An empty string is
// returned if the parameter is not found.
func Param(r *http.Request, key string) string {
	// Check if the params have been parsed for this request.
	params, ok := r.Context().Value(contextKeyParams).(map[string]string)
	if ok {
		return params[key]
	}

	// Get the param indexes to be able to parse the parameters from the path
	// for this request.
	paramsIdx, ok := r.Context().Value(contextKeyParamsIdx).(map[string]uint16)
	if !ok {
		return ""
	}

	params = make(map[string]string)
	parts := pathSegments(r.URL.Path)

	for name, idx := range paramsIdx {
		params[name] = parts[idx]
	}

	// Add the parsed parameter key/values into the request context.
	*r = *r.WithContext(context.WithValue(r.Context(), contextKeyParams, params))

	return params[key]
}
