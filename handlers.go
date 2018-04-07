package rest

import "net/http"

/*
Get handles the GET verb for individual items. Should be mapped to:
GET /thing/:id
*/
func Get(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.Get(w, r)
	}
}

/*
GetAll handles the GET verb for the full collection. Should be mapped to:
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
Post handles the POST verb. Should be mapped to:
POST /thing
*/
func Post(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.Post(w, r)
	}
}

/*
Put handles the PUT verb. Should be mapped to:
PUT /thing/:id
*/
func Put(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.Put(w, r)
	}
}

/*
Delete handles the DELETE verb. Should be mapped to:
DELETE /thing/:id
*/
func Delete(newRepository RepositoryConstructor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Controller{Repository: newRepository(r.Context())}
		c.Delete(w, r)
	}
}
