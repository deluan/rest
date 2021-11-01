package examples

import (
	"context"
	"errors"
	"strconv"

	"github.com/deluan/rest"
)

// ***********************************
// * Sample Repository and Model
// ***********************************

type ReadWriteRepository interface {
	rest.Repository
	rest.Persistable
}

// NewSampleRepository returns a new SampleRepository
func NewSampleRepository(ctx context.Context, logger ...rest.Logger) *SampleRepository {
	repo := SampleRepository{Context: ctx}
	repo.data = make(map[string]SampleModel)
	return &repo
}

type SampleModel struct {
	ID   string
	Name string
	Age  int
}

// SampleRepository is a simple in-memory repository implementation. NOTE: This repository does not handle QueryOptions
type SampleRepository struct {
	Context context.Context
	Error   error
	data    map[string]SampleModel
	seq     int64
}

func (r *SampleRepository) Count(options ...rest.QueryOptions) (int64, error) {
	return int64(len(r.data)), r.Error
}

func (r *SampleRepository) Read(id string) (interface{}, error) {
	if r.Error != nil {
		return nil, r.Error
	}
	if data, ok := r.data[id]; ok {
		return data, nil
	}
	return nil, rest.ErrNotFound
}

func (r *SampleRepository) ReadAll(options ...rest.QueryOptions) (interface{}, error) {
	if r.Error != nil {
		return nil, r.Error
	}
	dataSet := make([]SampleModel, 0)
	for _, v := range r.data {
		dataSet = append(dataSet, v)
	}
	return dataSet, nil
}

func (r *SampleRepository) EntityName() string {
	return "sample"
}

func (r *SampleRepository) NewInstance() interface{} {
	return &SampleModel{}
}

func NewPersistableSampleRepository(ctx context.Context, logger ...rest.Logger) *PersistableSampleRepository {
	repo := PersistableSampleRepository{}
	repo.Context = ctx
	repo.data = make(map[string]SampleModel)
	return &repo
}

type PersistableSampleRepository struct {
	SampleRepository
}

func (r *PersistableSampleRepository) Save(entity interface{}) (string, error) {
	if r.Error != nil {
		return "", r.Error
	}
	rec := entity.(*SampleModel)
	r.seq = r.seq + 1
	rec.ID = strconv.FormatInt(r.seq, 10)
	if _, ok := r.data[rec.ID]; ok {
		return "", errors.New("record already exists")
	}

	r.data[rec.ID] = *rec
	return rec.ID, nil
}

func (r *PersistableSampleRepository) Update(id string, entity interface{}, cols ...string) error {
	if r.Error != nil {
		return r.Error
	}
	rec := entity.(*SampleModel)
	if _, ok := r.data[rec.ID]; !ok {
		return rest.ErrNotFound
	}

	r.data[rec.ID] = *rec
	return nil
}

func (r *PersistableSampleRepository) Delete(id string) error {
	if r.Error != nil {
		return r.Error
	}
	if _, ok := r.data[id]; !ok {
		return rest.ErrNotFound
	}

	delete(r.data, id)
	return nil
}
