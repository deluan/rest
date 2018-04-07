# Simple Generic REST server

[![Build Status](https://travis-ci.org/deluan/rest.svg?branch=master)](https://travis-ci.org/deluan/rest) [![Go Report Card](https://goreportcard.com/badge/github.com/deluan/rest)](https://goreportcard.com/report/github.com/deluan/rest) 

Package rest provides a simple REST controller compatible with the [JSON Server API](https://github.com/typicode/json-server) 
"dialect". This package enables the creation of backends for the great [Admin-on-rest](https://marmelab.com/admin-on-rest/) 
framework using pure Go, but can be used in other scenarios where you need a simple REST server for your data.

To use it, you will need to provide an implementation of the Repository interface and a function to create
such repository.

The controller was created to be used with Gorilla Pat, as it requires URL params to be parsed and set
as query params. You can easily adapt it to work with other routers and frameworks using a custom middleware.

Example using [Gorilla Pat](https://github.com/gorilla/pat):

```go
	func NewThingsRepository(ctx context) rest.Repository {
		return &ThingsRepository{ctx: ctx}
	}

	func main() {
		router := pat.New()

		router.Get("/thing/{id}", rest.Get(NewThingsRepository))
		router.Get("/thing", rest.GetAll(NewThingsRepository))
		router.Post("/thing", rest.Post(NewThingsRepository))
		router.Put("/thing/{id}", rest.Put(NewThingsRepository))
		router.Delete("/thing/{id}", rest.Delete(NewThingsRepository))

		http.Handle("/", router)

		log.Print("Listening on 127.0.0.1:8000...")
		log.Fatal(http.ListenAndServe(":8000", nil))
	}
```

Example using [chi router](https://github.com/go-chi/chi):

```go
	func main() {
		router := chi.NewRouter()

		router.Route("/thing", func(r chi.Router) {
			r.Get("/", rest.GetAll(NewThingsRepository))
			r.Post("/", rest.Post(NewThingsRepository))
			r.Route("/{id:[0-9]+}", func(r chi.Router) {
				r.With(urlParams).Get("/", rest.Get(NewThingsRepository))
				r.With(urlParams).Put("/", rest.Put(NewThingsRepository))
				r.With(urlParams).Delete("/", rest.Delete(NewThingsRepository))
			})
		})

		http.Handle("/", router)

		log.Print("Listening on 127.0.0.1:8000...")
		log.Fatal(http.ListenAndServe(":8000", nil))
	}

	// Middleware to convert Chi URL params (from Context) to query params, as expected by our REST package
	func urlParams(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := chi.RouteContext(r.Context())
			parts := make([]string, 0)
			for i, key := range ctx.URLParams.Keys {
				value := ctx.URLParams.Values[i]
				if key == "*" {
					continue
				}
				parts = append(parts, url.QueryEscape(":"+key)+"="+url.QueryEscape(value))
			}
			q := strings.Join(parts, "&")
			if r.URL.RawQuery == "" {
				r.URL.RawQuery = q
			} else {
				r.URL.RawQuery += "&" + q
			}

			next.ServeHTTP(w, r)
		})
	}
```
