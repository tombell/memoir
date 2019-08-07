# journey

Simple, and opinionated HTTP router for Go.

## Rationale

There are many different HTTP routers and request multiplexers available for Go,
so why did I decide to write another?

I wanted to learn about [radix trees](https://en.wikipedia.org/wiki/Radix_tree),
which are utilised in a number of Go-based router libraries for performance, and
I'm very opinionated when it comes to what I'm looking for in a router package.

If you're looking for a very basic, but great solution, I recommend Mat Ryer's
[way](https://github.com/matryer/way) package.

## Features

- Simple and fast (utilise radix trees for route lookup performance)
- Route parameters, and optional regular expressions for parameters
- Wildcard route matching

## Usage

Import the package `github.com/tombell/journey` using your choice of dependency
management.

### Creating a Router

Create a new `Router` to begin registering routes and handlers.

```go
func main() {
	r := journey.NewRouter()
}
```

### Adding Routes and Handlers

Once you have a `Router` instance, you can begin adding your routes and
handlers.

```go
func main() {
	r := journey.NewRouter()

	r.HandleFunc("GET", "/songs", handleSongList)
	r.HandleFunc("POST", "/songs", handleSongAdd)
}

func handleSongList(w http.ResponseWriter, r *http.Request) {
	// ...
}

// ...
```

#### Routes with Parameters

You can include path parameters in your route patterns, and those are then
available inside the handler by using `journey.Param()`. A parameter inside the
route pattern is prefixed with a `:`.

```go
func main() {
	r := journey.NewRouter()

	// ...

	r.HandleFunc("GET", "/songs/:id", handleSongGet)
}

// ...

func handleSongGet(w http.ResponseWriter, r *http.Request) {
	id := journey.Param(r, "id")
	// ...
}
```

#### Regular Expression Constraints

You are able to specify regular expressions as part of a route parameter to add
a constraint. To add a regular expression constraint, after the parameter add
another `:` followed by the regular expression.

```go
func main() {
	r := journey.NewRouter()

	// ...

	r.HandleFunc("GET", `/songs/:id:^\d+$`, handleSongGet)
}
```

This means `/songs/1234` will match this route, however `/songs/abcd` will not.

#### Wildcard Matching

You can specify a route suffixed with `...` to allow it to match any path with
the prefix.

```go
func main() {
	r := journey.NewRouter()

	r.HandleFunc("GET", "/images...", handleImagesList)
}
```

The above pattern will match the following paths:

- `/images`
- `/images/`
- `/images/my/cool/image.png`

### Not Found Handler

You can specify a handler to use when a path is not matched, set the
`NotFoundHandler` field on the `Router`.

```go
func main() {
	r := journey.NewRouter()

	r.NotFoundHandler = http.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "This is not the page you are looking for")
	})
}
```

### Using the Router

Once you have all your routes registered, you can use the `Router` to listen on
for requests.

```go
func main() {
	r := journey.NewRouter()

	r.HandleFunc("GET", "/songs", handleSongList)
	r.HandleFunc("POST", "/songs", handleSongAdd)
	r.HandleFunc("GET", "/songs/:id", handleSongGet)

	http.ListenAndServe(":8080", r)
}
```
