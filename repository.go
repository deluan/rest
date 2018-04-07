package rest

import (
	"context"
	"errors"
)

/*
Possible errors returned by a Repository implementation. Any error other than these will make the REST controller
return a 500 http status code.
*/
var (
	// Will make the controller return a 404 error
	ErrNotFound = errors.New("data not found")

	// Will make the controller return a 403 error
	ErrPermissionDenied = errors.New("permission denied")
)

/*
QueryOptions are optional query parameters that can be received by Count and ReadAll and are used to implement pagination,
sorting and filtering.
*/
type QueryOptions struct {
	// Comma separated list of fields to sort the data
	Sort string

	// Possible values: asc (default), desc
	Order string

	// Max records to return. Used for pagination
	Max int

	// Initial record to return. Used for pagination
	Offset int

	// Map of filters to apply to the query. Keys are field names and values are the filter
	// to be applied to the field. Eg.: {"age": 30, "name": "john"}
	// How the values of the filters are applied to the fields is implementation dependent
	// (you can implement substring, exact match, etc..)
	Filters map[string]interface{}
}

/*
RepositoryConstructor needs to be implemented by your custom repository implementation, and it returns a fully
initialized repository. It is meant to be called on every HTTP request, so you shouldn't keep state in your repository,
and it should execute fast. You have access to the current HTTP request's context.
*/
type RepositoryConstructor func(ctx context.Context) Repository

/*
Repository is the interface that must be created for your data. See SampleRepository (in examples folder) for a simple
in-memory map-based example.
*/
type Repository interface {
	// Returns the number of entities that matches the criteria specified by the options
	Count(options ...QueryOptions) (int64, error)

	// Returns the entity identified by id
	Read(id int64) (interface{}, error)

	// Returns a slice of entities that matches the criteria specified by the options
	ReadAll(options ...QueryOptions) (interface{}, error)

	// Adds the entity to the repository and returns the newly created id
	Save(entity interface{}) (int64, error)

	// Updates the entity identified by id. Optionally select the fields to be updated
	Update(entity interface{}, cols ...string) error

	// Delete the entity identified by id
	Delete(id int64) error

	// Return the entity name (used for logs and messages)
	EntityName() string

	// Returns a newly created instance. Should be as simple as return &Thing{}
	NewInstance() interface{}
}
