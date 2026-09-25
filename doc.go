/*
Package rest provides a simple REST controller compatible with the JSON Server API "dialect". This package enables the
creation of backends for the great React-admin framework using pure Go, but can be used in other scenarios where you
need a simple REST server for your data.

To use it, you need to provide an implementation of the generic Repository[T] interface for your entity type T.
Implement Persistable[T] as well to enable the POST, PUT and DELETE methods; without it, those methods return
405 Method Not Allowed. Every repository method receives the request's context.Context as its first parameter.

The controller was created to be used with Gorilla Pat, as it requires URL params to be parsed and set
as query params. You can easily adapt it to work with other routers and frameworks using a custom middleware.

The functionality is provided by a set of handlers named after the REST verbs they handle: Get(), GetAll(), Put(),
Post() and Delete(). Each of these functions receives your repository instance and returns an http.HandlerFunc.

Errors returned by your repository are matched with errors.Is / errors.As, so they can be wrapped. ErrNotFound
maps to 404 (for example {"error":"Thing(id:1) not found"}), ErrPermissionDenied to 403, a *ValidationError to 400
with the field errors in the body, and any other error to 500.

Example using Gorilla Pat (https://github.com/gorilla/pat):

	type Thing struct {
		ID   string
		Name string
	}

	// ThingsRepository implements rest.Repository[Thing] and rest.Persistable[Thing]
	type ThingsRepository struct {
		// your storage, e.g. a *sql.DB
	}

	func main() {
		repo := rest.Repository[Thing](&ThingsRepository{})
		router := pat.New()

		router.Get("/thing/{id}", rest.Get(repo))
		router.Get("/thing", rest.GetAll(repo))
		router.Post("/thing", rest.Post(repo))
		router.Put("/thing/{id}", rest.Put(repo))
		router.Delete("/thing/{id}", rest.Delete(repo))

		http.Handle("/", router)

		log.Print("Listening on 127.0.0.1:8000...")
		log.Fatal(http.ListenAndServe(":8000", nil))
	}

Example using chi router (https://github.com/go-chi/chi):

	func main() {
		repo := rest.Repository[Thing](&ThingsRepository{})
		router := chi.NewRouter()

		router.Route("/thing", func(r chi.Router) {
			r.Get("/", rest.GetAll(repo))
			r.Post("/", rest.Post(repo))
			r.Route("/{id:[0-9]+}", func(r chi.Router) {
				r.With(urlParams).Get("/", rest.Get(repo))
				r.With(urlParams).Put("/", rest.Put(repo))
				r.With(urlParams).Delete("/", rest.Delete(repo))
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

For more info see:

	JSON Server: https://github.com/typicode/json-server
	React-admin: https://marmelab.com/react-admin/
*/
package rest
