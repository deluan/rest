package rest

import (
	"net/http"
)

/*
Get handles the GET verb for individual items. Should be mapped to:
GET /thing/:id
*/
func Get[T any](repository Repository[T]) http.HandlerFunc {
	c := createController(repository)
	return func(w http.ResponseWriter, r *http.Request) {
		c.Get(w, r)
	}
}

/*
GetAll handles the GET verb for the full collection. Should be mapped to:
GET /thing
For all query options available, see https://github.com/typicode/json-server
*/
func GetAll[T any](repository Repository[T]) http.HandlerFunc {
	c := createController(repository)
	return func(w http.ResponseWriter, r *http.Request) {
		c.GetAll(w, r)
	}
}

/*
Post handles the POST verb. Should be mapped to:
POST /thing
*/
func Post[T any](repository Repository[T]) http.HandlerFunc {
	c := createController(repository)
	return func(w http.ResponseWriter, r *http.Request) {
		c.Post(w, r)
	}
}

/*
Put handles the PUT verb. Should be mapped to:
PUT /thing/:id
*/
func Put[T any](repository Repository[T]) http.HandlerFunc {
	c := createController(repository)
	return func(w http.ResponseWriter, r *http.Request) {
		c.Put(w, r)
	}
}

/*
Delete handles the DELETE verb. Should be mapped to:
DELETE /thing/:id
*/
func Delete[T any](repository Repository[T]) http.HandlerFunc {
	c := createController(repository)
	return func(w http.ResponseWriter, r *http.Request) {
		c.Delete(w, r)
	}
}

func createController[T any](r Repository[T]) *Controller[T] {
	c := &Controller[T]{Repository: r}
	return c
}
