package rest

import (
	"context"
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
	// to be applied to the field. E.g.: {"age": 30, "name": "john"}.
	// How the values of the filters are applied to the fields is implementation dependent
	// (you can implement substring, exact match, etc...)
	Filters map[string]any
}

/*
Repository is the interface that must be created to access your data. This will be used for the GET http method.
*/
type Repository[T any] interface {
	// Count returns the number of entities that matches the criteria specified by the options
	Count(context.Context, ...QueryOptions) (int64, error)

	// Read returns the entity identified by id
	Read(ctx context.Context, id string) (*T, error)

	// ReadAll returns a slice of entities that matches the criteria specified by the options
	ReadAll(context.Context, ...QueryOptions) ([]T, error)
}

/*
Persistable must be implemented by repositories in addition to the Repository interface, to allow the POST,
PUT and DELETE http methods. If this interface is not implemented by the repository, calls to these http methods
will return 405 - Method Not Allowed.
*/
type Persistable[T any] interface {
	Repository[T]

	// Save adds the entity to the repository and returns the newly created id
	Save(ctx context.Context, entity *T) (string, error)

	// Update the entity identified by id. Optionally select the fields to be updated
	Update(ctx context.Context, id string, entity T, fields ...string) error

	// Delete the entities identified by id
	Delete(ctx context.Context, ids ...string) error
}
