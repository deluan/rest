package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestController_GetAll(t *testing.T) {
	Convey("Given an empty repository", t, func() {
		repo := NewFakeRepository()
		controller := &Controller{repo}

		Convey("When I call GetAll", func() {
			req := httptest.NewRequest("GET", "/fake", nil)
			res := httptest.NewRecorder()
			controller.GetAll(res, req)

			Convey("It returns 200 http status", func() {
				So(res.Code, ShouldEqual, 200)
			})

			Convey("It returns an empty collection", func() {
				So(res.Body.String(), ShouldEqual, "[]")
			})

			Convey("It returns 0 in the X-Total-Count header", func() {
				So(res.HeaderMap["X-Total-Count"][0], ShouldEqual, "0")
			})
		})

		Convey("When two items are added", func() {
			joe := aRecord("Joe", 30)
			idJoe, _ := repo.Save(&joe)
			cecilia := aRecord("Cecilia", 22)
			idCecilia, _ := repo.Save(&cecilia)

			Convey("And I call GetAll", func() {
				req := httptest.NewRequest("GET", "/fake", nil)
				res := httptest.NewRecorder()
				controller.GetAll(res, req)

				Convey("It returns 200 http status", func() {
					So(res.Code, ShouldEqual, 200)
				})

				Convey("It returns 2 records", func() {
					response := make([]FakeModel, 0)
					if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
						panic(err)
					}
					So(response, ShouldHaveLength, 2)
					for _, record := range response {
						switch record.Id {
						case idJoe:
							So(record.Name, ShouldEqual, "Joe")
							So(record.Age, ShouldEqual, 30)
						case idCecilia:
							So(record.Name, ShouldEqual, "Cecilia")
							So(record.Age, ShouldEqual, 22)
						default:
							t.Errorf("Invalid record returned: %#v", record)
						}
					}
				})

				Convey("It returns 2 in the X-Total-Count header", func() {
					So(res.HeaderMap["X-Total-Count"][0], ShouldEqual, "2")
				})
			})
		})

		Convey("When the repository returns an error", func() {
			repo.err = errors.New("unknown error")
			req := httptest.NewRequest("GET", "/fake", nil)
			res := httptest.NewRecorder()
			controller.GetAll(res, req)

			Convey("It returns 500 http status", func() {
				So(res.Code, ShouldEqual, 500)
			})
		})
	})
}

func TestController_Get(t *testing.T) {
	Convey("Given an empty repository", t, func() {
		repo := NewFakeRepository()
		controller := &Controller{repo}

		Convey("When I call Get id=1", func() {
			req := httptest.NewRequest("GET", "/fake?:id=1", nil)
			res := httptest.NewRecorder()
			controller.Get(res, req)

			Convey("It returns 404 http status", func() {
				So(res.Code, ShouldEqual, 404)
			})

			Convey("It returns an error message in the response", func() {
				var response map[string]string
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "error")
			})
		})

		Convey("When an item is added", func() {
			joe := aRecord("Joe", 30)
			id, _ := repo.Save(&joe)

			Convey("And I call Get", func() {
				req := httptest.NewRequest("GET", fmt.Sprintf("/fake?:id=%d", id), nil)
				res := httptest.NewRecorder()
				controller.Get(res, req)

				Convey("It returns 200 http status", func() {
					So(res.Code, ShouldEqual, 200)
				})

				Convey("It returns all data from the record", func() {
					var response FakeModel
					if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
						panic(err)
					}
					So(response.Id, ShouldEqual, id)
					So(response.Name, ShouldEqual, "Joe")
					So(response.Age, ShouldEqual, 30)
				})
			})
		})

		Convey("When the repository returns an error", func() {
			repo.err = errors.New("unknown error")
			req := httptest.NewRequest("GET", "/fake?:id=1", nil)
			res := httptest.NewRecorder()
			controller.Get(res, req)

			Convey("It returns 500 http status", func() {
				So(res.Code, ShouldEqual, 500)
			})

			Convey("It returns an error message in the response", func() {
				var response map[string]string
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "error")
			})
		})
	})
}

func TestController_Delete(t *testing.T) {
	Convey("Given an empty repository", t, func() {
		repo := NewFakeRepository()
		controller := &Controller{repo}

		Convey("When I call Delete id=1", func() {
			req := httptest.NewRequest("DELETE", "/fake?:id=1", nil)
			res := httptest.NewRecorder()
			controller.Delete(res, req)

			Convey("It returns 404 http status", func() {
				So(res.Code, ShouldEqual, 404)
			})

			Convey("It returns an error message in the response", func() {
				var response map[string]string
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "error")
			})
		})

		Convey("When the repository returns an error", func() {
			repo.err = errors.New("unknown error")

			req := httptest.NewRequest("DELETE", "/fake?:id=1", aRecordReader("John Doe", 33))
			res := httptest.NewRecorder()
			controller.Delete(res, req)

			Convey("It returns 500 http status", func() {
				So(res.Code, ShouldEqual, 500)
			})

			Convey("It returns an error message in the response", func() {
				var response map[string]string
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "error")
			})
		})
	})
}

func TestController_Put(t *testing.T) {
	Convey("Given an empty repository", t, func() {
		repo := NewFakeRepository()
		controller := &Controller{repo}

		Convey("When I call Put with an invalid request", func() {
			req := httptest.NewRequest("PUT", "/fake?:id=1", nil)
			res := httptest.NewRecorder()
			controller.Put(res, req)

			Convey("It returns 422 http status", func() {
				So(res.Code, ShouldEqual, 422)
			})

			Convey("It returns an error message in the response", func() {
				var response map[string]string
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "error")
			})
		})

		Convey("When I call Put id=1", func() {
			req := httptest.NewRequest("PUT", "/fake?:id=1", aRecordReader("John Doe", 33))
			res := httptest.NewRecorder()
			controller.Put(res, req)

			Convey("It returns 404 http status", func() {
				So(res.Code, ShouldEqual, 404)
			})

			Convey("It returns an error message in the response", func() {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "error")
			})
		})

		Convey("When the repository returns an error", func() {
			repo.err = errors.New("unknown error")
			req := httptest.NewRequest("PUT", "/fake?:id=1", aRecordReader("John Doe", 33))
			res := httptest.NewRecorder()
			controller.Put(res, req)

			Convey("It returns 500 http status", func() {
				So(res.Code, ShouldEqual, 500)
			})

			Convey("It returns an error message in the response", func() {
				var response map[string]string
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "error")
			})
		})
	})
}

func TestController_Post(t *testing.T) {
	Convey("Given an empty repository", t, func() {
		repo := NewFakeRepository()
		controller := &Controller{repo}

		Convey("When I send valid data", func() {
			req := httptest.NewRequest("POST", "/fake", aRecordReader("John Doe", 33))
			res := httptest.NewRecorder()
			controller.Post(res, req)

			Convey("It returns 200 http status", func() {
				So(res.Code, ShouldEqual, 200)
			})

			Convey("It adds the data to the repo", func() {
				count, _ := repo.Count()
				So(count, ShouldEqual, 1)
			})

			Convey("It returns the new id in the response", func() {
				var response map[string]int64
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "id")
				id := response["id"]
				r, _ := repo.Read(id)
				So(r.(FakeModel).Name, ShouldEqual, "John Doe")
			})
		})

		Convey("When I send invalid data", func() {
			req := httptest.NewRequest("POST", "/fake", strings.NewReader("BAD DATA"))
			res := httptest.NewRecorder()
			controller.Post(res, req)

			Convey("It returns 422 http status", func() {
				So(res.Code, ShouldEqual, 422)
			})

			Convey("It does not adds any data to the repo", func() {
				count, _ := repo.Count()
				So(count, ShouldEqual, 0)
			})
		})

		Convey("When the repository returns an error", func() {
			repo.err = errors.New("unknown error")

			req := httptest.NewRequest("POST", "/fake?:id=1", aRecordReader("John Doe", 33))
			res := httptest.NewRecorder()
			controller.Post(res, req)

			Convey("It returns 500 http status", func() {
				So(res.Code, ShouldEqual, 500)
			})

			Convey("It returns an error message in the response", func() {
				var response map[string]string
				if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
					panic(err)
				}
				So(response, ShouldContainKey, "error")
			})
		})
	})
}

func aRecord(name string, age int) FakeModel {
	return FakeModel{Name: name, Age: age}
}

func aRecordReader(name string, age int) io.Reader {
	r := aRecord(name, age)
	buf, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(buf)
}

// ***********************************
// * Fake Repository and Model
// ***********************************

func NewFakeRepository() *FakeRepository {
	repo := FakeRepository{}
	repo.data = make(map[int64]FakeModel)
	return &repo
}

type FakeModel struct {
	Id   int64
	Name string
	Age  int
}

type FakeRepository struct {
	data map[int64]FakeModel
	err  error
	seq  int64
}

func (r *FakeRepository) Count(options ...QueryOptions) (int64, error) {
	return int64(len(r.data)), r.err
}
func (r *FakeRepository) Read(id int64) (interface{}, error) {
	if r.err != nil {
		return nil, r.err
	}
	if data, ok := r.data[id]; ok {
		return data, nil
	}
	return nil, ErrNotFound
}
func (r *FakeRepository) ReadAll(options ...QueryOptions) (interface{}, error) {
	if r.err != nil {
		return nil, r.err
	}
	dataSet := make([]FakeModel, 0)
	for _, v := range r.data {
		dataSet = append(dataSet, v)
	}
	return dataSet, nil
}
func (r *FakeRepository) Save(entity interface{}) (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	rec := entity.(*FakeModel)
	r.seq = r.seq + 1
	rec.Id = r.seq
	if _, ok := r.data[rec.Id]; ok {
		return -1, errors.New("record already exists")
	}

	r.data[rec.Id] = *rec
	return rec.Id, nil
}
func (r *FakeRepository) Update(entity interface{}, cols ...string) error {
	if r.err != nil {
		return r.err
	}
	rec := entity.(*FakeModel)
	if _, ok := r.data[rec.Id]; !ok {
		return ErrNotFound
	}

	r.data[rec.Id] = *rec
	return nil
}
func (r *FakeRepository) Delete(id int64) error {
	if r.err != nil {
		return r.err
	}
	if _, ok := r.data[id]; !ok {
		return ErrNotFound
	}

	delete(r.data, id)
	return nil
}
func (r *FakeRepository) EntityName() string {
	return "fake"
}
func (r *FakeRepository) NewInstance() interface{} {
	return &FakeModel{}
}

func init() {
	log.SetLevel(log.FatalLevel)
}
