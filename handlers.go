package rest

import "net/http"

/*
Handles the GET verb for individual items. Should be mapped to:
GET /thing/:id
*/
func Get(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.Get(w, r)
	}
}

/*
Handles the GET verb for the full collection. Should be mapped to:
GET /thing
For all query options available, see https://github.com/typicode/json-server
*/
func GetAll(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.GetAll(w, r)
	}
}

/*
Handles the POST verb. Should be mapped to:
POST /thing
*/
func Post(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.Post(w, r)
	}
}

/*
Handles the PUT verb. Should be mapped to:
PUT /thing/:id
*/
func Put(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.Put(w, r)
	}
}

/*
Handles the DELETE verb. Should be mapped to:
DELETE /thing/:id
*/
func Delete(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.Delete(w, r)
	}
}
