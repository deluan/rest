package rest

import (
	"context"
	"net/http"
)

/*
Get handles the GET verb for individual items. Should be mapped to:
GET /thing/:id
*/
func Get(newRepository RepositoryConstructor, logger ...Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := createController(newRepository, r.Context(), logger...)
		c.Get(w, r)
	}
}

/*
GetAll handles the GET verb for the full collection. Should be mapped to:
GET /thing
For all query options available, see https://github.com/typicode/json-server
*/
func GetAll(newRepository RepositoryConstructor, logger ...Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := createController(newRepository, r.Context(), logger...)
		c.GetAll(w, r)
	}
}

/*
Post handles the POST verb. Should be mapped to:
POST /thing
*/
func Post(newRepository RepositoryConstructor, logger ...Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := createController(newRepository, r.Context(), logger...)
		c.Post(w, r)
	}
}

/*
Put handles the PUT verb. Should be mapped to:
PUT /thing/:id
*/
func Put(newRepository RepositoryConstructor, logger ...Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := createController(newRepository, r.Context(), logger...)
		c.Put(w, r)
	}
}

/*
Delete handles the DELETE verb. Should be mapped to:
DELETE /thing/:id
*/
func Delete(newRepository RepositoryConstructor, logger ...Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := createController(newRepository, r.Context(), logger...)
		c.Delete(w, r)
	}
}

func createController(newRepository RepositoryConstructor, ctx context.Context, logger ...Logger) Controller {
	c := Controller{Repository: newRepository(ctx)}
	if len(logger) > 0 {
		c.Logger = logger[0]
	}
	return c
}