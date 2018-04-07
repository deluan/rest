package rest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/deluan/rest"
	"github.com/deluan/rest/examples"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

var logger = logrus.New()

func TestController_GetAll(t *testing.T) {
	Convey("Given an empty repository", t, func() {
		repo := examples.NewSampleRepository(nil)
		controller := &rest.Controller{Repository: repo, Logger: logger}

		Convey("When I call GetAll", func() {
			req := httptest.NewRequest("GET", "/sample", nil)
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
				req := httptest.NewRequest("GET", "/sample", nil)
				res := httptest.NewRecorder()
				controller.GetAll(res, req)

				Convey("It returns 200 http status", func() {
					So(res.Code, ShouldEqual, 200)
				})

				Convey("It returns 2 records", func() {
					response := make([]examples.SampleModel, 0)
					if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
						panic(err)
					}
					So(response, ShouldHaveLength, 2)
					for _, record := range response {
						switch record.ID {
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
			repo.(*examples.SampleRepository).Error = errors.New("unknown error")
			req := httptest.NewRequest("GET", "/sample", nil)
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
		repo := examples.NewSampleRepository(nil)
		controller := &rest.Controller{Repository: repo, Logger: logger}

		Convey("When I call Get id=1", func() {
			req := httptest.NewRequest("GET", "/sample?:id=1", nil)
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
				req := httptest.NewRequest("GET", fmt.Sprintf("/sample?:id=%d", id), nil)
				res := httptest.NewRecorder()
				controller.Get(res, req)

				Convey("It returns 200 http status", func() {
					So(res.Code, ShouldEqual, 200)
				})

				Convey("It returns all data from the record", func() {
					var response examples.SampleModel
					if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
						panic(err)
					}
					So(response.ID, ShouldEqual, id)
					So(response.Name, ShouldEqual, "Joe")
					So(response.Age, ShouldEqual, 30)
				})
			})
		})

		Convey("When the repository returns an error", func() {
			repo.(*examples.SampleRepository).Error = errors.New("unknown error")
			req := httptest.NewRequest("GET", "/sample?:id=1", nil)
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
		repo := examples.NewSampleRepository(nil)
		controller := &rest.Controller{repo, logger}

		Convey("When I call Delete id=1", func() {
			req := httptest.NewRequest("DELETE", "/sample?:id=1", nil)
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

		Convey("When an item is added", func() {
			joe := aRecord("Joe", 30)
			id, err := repo.Save(&joe)

			So(err, ShouldBeNil)

			Convey("And I call Put", func() {
				req := httptest.NewRequest("DELETE", fmt.Sprintf("/sample?:id=%d", id), nil)
				res := httptest.NewRecorder()
				controller.Delete(res, req)

				Convey("It returns 200 http status", func() {
					So(res.Code, ShouldEqual, 200)
				})

				Convey("It deletes the record from the repository", func() {
					_, err := repo.Read(id)
					So(err, ShouldEqual, rest.ErrNotFound)
				})
			})
		})

		Convey("When the repository returns an error", func() {
			repo.(*examples.SampleRepository).Error = errors.New("unknown error")

			req := httptest.NewRequest("DELETE", "/sample?:id=1", nil)
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
		repo := examples.NewSampleRepository(nil)
		controller := &rest.Controller{repo, logger}

		Convey("When I call Put with an invalid request", func() {
			req := httptest.NewRequest("PUT", "/sample?:id=1", nil)
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
			req := httptest.NewRequest("PUT", "/sample", aRecordReader(1, "John Doe", 33))
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

		Convey("When an item is added", func() {
			joe := aRecord("Joe", 30)
			id, _ := repo.Save(&joe)

			Convey("And I call Put", func() {
				req := httptest.NewRequest("PUT", fmt.Sprintf("/sample?:id=%d", id), aRecordReader(id, "John", 31))
				res := httptest.NewRecorder()
				controller.Put(res, req)

				Convey("It returns 200 http status", func() {
					So(res.Code, ShouldEqual, 200)
				})

				Convey("It returns all data from the record", func() {
					var response examples.SampleModel
					if err := json.Unmarshal([]byte(res.Body.String()), &response); err != nil {
						panic(err)
					}
					So(response.ID, ShouldEqual, id)
					So(response.Name, ShouldEqual, "John")
					So(response.Age, ShouldEqual, 31)
				})
			})
		})

		Convey("When the repository returns an error", func() {
			repo.(*examples.SampleRepository).Error = errors.New("unknown error")
			req := httptest.NewRequest("PUT", "/sample", aRecordReader(1, "John Doe", 33))
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
		repo := examples.NewSampleRepository(nil)
		controller := &rest.Controller{repo, logger}

		Convey("When I send valid data", func() {
			req := httptest.NewRequest("POST", "/sample", aRecordReader(0, "John Doe", 33))
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
				So(r.(examples.SampleModel).Name, ShouldEqual, "John Doe")
			})
		})

		Convey("When I send invalid data", func() {
			req := httptest.NewRequest("POST", "/sample", strings.NewReader("BAD DATA"))
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
			repo.(*examples.SampleRepository).Error = errors.New("unknown error")

			req := httptest.NewRequest("POST", "/sample", aRecordReader(0, "John Doe", 33))
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

func aRecord(name string, age int) examples.SampleModel {
	return examples.SampleModel{Name: name, Age: age}
}

func aRecordReader(id int64, name string, age int) io.Reader {
	r := aRecord(name, age)
	r.ID = id
	buf, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(buf)
}

func init() {
	logger.SetLevel(logrus.FatalLevel)
}
