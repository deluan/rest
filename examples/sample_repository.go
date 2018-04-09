package examples

import (
	"context"
	"errors"

	"github.com/deluan/rest"
)

// ***********************************
// * Sample Repository and Model
// ***********************************

// SampleRepository Constructor
func NewSampleRepository(ctx context.Context, logger ...rest.Logger) rest.Repository {
	repo := SampleRepository{Context: ctx}
	repo.data = make(map[int64]SampleModel)
	return &repo
}

type SampleModel struct {
	ID   int64
	Name string
	Age  int
}

// Simple in-memory repository implementation. NOTE: This repository does not handle QueryOptions
type SampleRepository struct {
	Context context.Context
	Error   error
	data    map[int64]SampleModel
	seq     int64
}

func (r *SampleRepository) Count(options ...rest.QueryOptions) (int64, error) {
	return int64(len(r.data)), r.Error
}

func (r *SampleRepository) Read(id int64) (interface{}, error) {
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

func (r *SampleRepository) Save(entity interface{}) (int64, error) {
	if r.Error != nil {
		return 0, r.Error
	}
	rec := entity.(*SampleModel)
	r.seq = r.seq + 1
	rec.ID = r.seq
	if _, ok := r.data[rec.ID]; ok {
		return -1, errors.New("record already exists")
	}

	r.data[rec.ID] = *rec
	return rec.ID, nil
}

func (r *SampleRepository) Update(entity interface{}, cols ...string) error {
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

func (r *SampleRepository) Delete(id int64) error {
	if r.Error != nil {
		return r.Error
	}
	if _, ok := r.data[id]; !ok {
		return rest.ErrNotFound
	}

	delete(r.data, id)
	return nil
}

func (r *SampleRepository) EntityName() string {
	return "sample"
}

func (r *SampleRepository) NewInstance() interface{} {
	return &SampleModel{}
}
