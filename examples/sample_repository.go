package examples

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/deluan/rest"
)

// ***********************************
// * Sample Repository and Model
// ***********************************

// NewSampleRepository returns a new SampleRepository
func NewSampleRepository() *SampleRepository {
	return &SampleRepository{}
}

type SampleModel struct {
	ID   string
	Name string
	Age  int
}

// SampleRepository is a simple in-memory repository implementation. NOTE: This repository does not handle QueryOptions
type SampleRepository struct {
	err  atomic.Pointer[error]
	seq  atomic.Int64
	data sync.Map
}

// SetError simulates an error by forcing all methods to return the specified error
func (r *SampleRepository) SetError(err error) {
	r.err.Store(&err)
}

// error is a helper method to simplify access to the err atomic value
func (r *SampleRepository) error() error {
	err := r.err.Load()
	if err != nil {
		return *err
	}
	return nil
}

func (r *SampleRepository) Count(_ context.Context, _ ...rest.QueryOptions) (int64, error) {
	count := 0
	r.data.Range(func(_, _ any) bool {
		count++
		return true
	})
	return int64(count), r.error()
}

func (r *SampleRepository) Read(_ context.Context, id string) (*SampleModel, error) {
	if err := r.error(); err != nil {
		return nil, err
	}
	if data, ok := r.data.Load(id); ok {
		entity := data.(SampleModel)
		return &entity, nil
	}
	return nil, rest.ErrNotFound
}

func (r *SampleRepository) ReadAll(_ context.Context, _ ...rest.QueryOptions) ([]SampleModel, error) {
	if err := r.error(); err != nil {
		return nil, err
	}
	dataSet := make([]SampleModel, 0)
	r.data.Range(func(_, v any) bool {
		dataSet = append(dataSet, v.(SampleModel))
		return true
	})
	return dataSet, nil
}

// NewPersistableSampleRepository returns a new PersistableSampleRepository
func NewPersistableSampleRepository() *PersistableSampleRepository {
	return &PersistableSampleRepository{}
}

// PersistableSampleRepository implements a read-write repository on top of the read-only SampleRepository
type PersistableSampleRepository struct {
	SampleRepository
}

func (r *PersistableSampleRepository) Save(_ context.Context, entity *SampleModel) (string, error) {
	if err := r.error(); err != nil {
		return "", err
	}
	entity.ID = strconv.FormatInt(r.seq.Add(1), 10)
	if _, loaded := r.data.LoadOrStore(entity.ID, *entity); loaded {
		return "", errors.New("record already exists")
	}
	return entity.ID, nil
}

func (r *PersistableSampleRepository) Update(_ context.Context, id string, entity SampleModel, cols ...string) error {
	if err := r.error(); err != nil {
		return err
	}
	data, ok := r.data.Load(id)
	if !ok {
		return rest.ErrNotFound
	}
	current := data.(SampleModel)
	if len(cols) == 0 {
		current = entity
		current.ID = id
	} else {
		for _, col := range cols {
			switch col {
			case "age":
				current.Age = entity.Age
			case "name":
				current.Name = entity.Name
			}
		}
	}
	r.data.Store(id, current)
	return nil
}

func (r *PersistableSampleRepository) Delete(_ context.Context, ids ...string) error {
	if err := r.error(); err != nil {
		return err
	}
	for _, id := range ids {
		if _, ok := r.data.Load(id); !ok {
			return rest.ErrNotFound
		}
	}

	for _, id := range ids {
		r.data.Delete(id)
	}
	return nil
}
