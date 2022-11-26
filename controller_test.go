package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/deluan/rest"
	"github.com/deluan/rest/examples"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func createRequestResponse(method, target string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, body)
	res := httptest.NewRecorder()
	return req, res
}

var _ = Describe("Handlers", func() {
	var handler http.HandlerFunc
	var repo *examples.PersistableSampleRepository
	var ctx = context.Background()

	BeforeEach(func() {
		repo = examples.NewPersistableSampleRepository()
	})

	Describe("GetAll", func() {
		var req *http.Request
		var res *httptest.ResponseRecorder
		BeforeEach(func() {
			handler = rest.GetAll(rest.Repository[examples.SampleModel](repo))
			req, res = createRequestResponse("GET", "/sample", nil)
			handler(res, req)
		})
		Context("Given an empty repository", func() {
			When("I call GetAll", func() {
				It("returns 200 http status", func() {
					Expect(res.Code).To(Equal(200))
				})
				It("returns an empty collection", func() {
					Expect(res.Body.String()).To(Equal("[]"))
				})
				It("returns 0 in the X-Total-Count header", func() {
					Expect(res.Header()["X-Total-Count"][0]).To(Equal("0"))
				})
			})
		})
		Context("Given a repository with two items added", func() {
			var idJoe, idCecilia string
			BeforeEach(func() {
				joe := aRecord("Joe", 30)
				idJoe, _ = repo.Save(ctx, &joe)
				cecilia := aRecord("Cecilia", 22)
				idCecilia, _ = repo.Save(ctx, &cecilia)
			})
			When("I call GetAll", func() {
				var req *http.Request
				var res *httptest.ResponseRecorder
				BeforeEach(func() {
					req, res = createRequestResponse("GET", "/sample", nil)
					handler(res, req)
				})
				It("returns 200 http status", func() {
					Expect(res.Code).To(Equal(200))
				})
				It("returns 2 records", func() {
					response := make([]examples.SampleModel, 0)
					err := json.Unmarshal(res.Body.Bytes(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(HaveLen(2))
					for _, record := range response {
						switch record.ID {
						case idJoe:
							Expect(record.Name).To(Equal("Joe"))
							Expect(record.Age).To(Equal(30))
						case idCecilia:
							Expect(record.Name).To(Equal("Cecilia"))
							Expect(record.Age).To(Equal(22))
						default:
							log.Panicf("Invalid record returned: %#v", record)
						}
					}
				})
				It("returns 2 in the X-Total-Count header", func() {
					Expect(res.Header()["X-Total-Count"][0]).To(Equal("2"))
				})
			})
		})
		Context("When the repository returns an ErrPermissionDenied", func() {
			It("returns 403 http status", func() {
				repo.SetError(rest.ErrPermissionDenied)
				req, res := createRequestResponse("GET", "/sample", nil)
				handler(res, req)

				Expect(res.Code).To(Equal(403))
			})
		})
		Context("When the repository returns an error", func() {
			It("returns 500 http status", func() {
				repo.SetError(errors.New("unknown error"))
				req, res := createRequestResponse("GET", "/sample", nil)
				handler(res, req)

				Expect(res.Code).To(Equal(500))
			})
		})
	})
	Describe("Get", func() {
		var req *http.Request
		var res *httptest.ResponseRecorder
		BeforeEach(func() {
			handler = rest.Get(rest.Repository[examples.SampleModel](repo))
		})
		Context("Given an empty repository", func() {
			When("I call GET with id=1", func() {
				BeforeEach(func() {
					req, res = createRequestResponse("GET", "/sample?:id=1", nil)
					handler(res, req)
				})
				It("returns 404 http status", func() {
					Expect(res.Code).To(Equal(404))
				})
				It("returns an error message in the response", func() {
					var response map[string]string
					err := json.Unmarshal(res.Body.Bytes(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(HaveKey("error"))
				})
			})
		})
		Context("Given a repository with one item", func() {
			var idJoe string
			BeforeEach(func() {
				joe := aRecord("Joe", 30)
				idJoe, _ = repo.Save(ctx, &joe)
			})
			When("I call GET with an existing id", func() {
				BeforeEach(func() {
					req, res = createRequestResponse("GET", fmt.Sprintf("/sample?:id=%s", idJoe), nil)
					handler(res, req)
				})
				It("returns 200 http status", func() {
					Expect(res.Code).To(Equal(200))
				})
				It("returns the complete record", func() {
					var response examples.SampleModel
					err := json.Unmarshal(res.Body.Bytes(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response.ID).To(Equal(idJoe))
					Expect(response.Name).To(Equal("Joe"))
					Expect(response.Age).To(Equal(30))
				})
			})
			Context("When the repository returns an ErrPermissionDenied", func() {
				It("returns 403 http status", func() {
					repo.SetError(rest.ErrPermissionDenied)
					req, res = createRequestResponse("GET", fmt.Sprintf("/sample?:id=%s", idJoe), nil)
					handler(res, req)

					Expect(res.Code).To(Equal(403))
				})
			})
			Context("When the repository returns an error", func() {
				BeforeEach(func() {
					repo.SetError(errors.New("unknown error"))
					req, res = createRequestResponse("GET", fmt.Sprintf("/sample?:id=%s", idJoe), nil)
					handler(res, req)
				})
				It("returns 500 http status", func() {
					Expect(res.Code).To(Equal(500))
				})
				It("returns an error message in the response", func() {
					var response map[string]string
					err := json.Unmarshal(res.Body.Bytes(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(HaveKey("error"))
				})
			})
		})
	})
	Describe("Delete", func() {
		var req *http.Request
		var res *httptest.ResponseRecorder
		Context("Given a read-only repository", func() {
			var repo *examples.SampleRepository
			BeforeEach(func() {
				repo = examples.NewSampleRepository()
				handler = rest.Delete(rest.Repository[examples.SampleModel](repo))
			})
			When("I call DELETE id=1", func() {
				BeforeEach(func() {
					req, res = createRequestResponse("DELETE", "/sample?:id=1", nil)
					handler(res, req)
				})
				It("returns 405 http status", func() {
					Expect(res.Code).To(Equal(405))
				})
			})
		})
		Context("With a persistable repository", func() {
			var repo *examples.PersistableSampleRepository
			BeforeEach(func() {
				repo = examples.NewPersistableSampleRepository()
				handler = rest.Delete(rest.Repository[examples.SampleModel](repo))
			})
			Context("Given an empty repository", func() {
				When("I call DELETE id=1", func() {
					BeforeEach(func() {
						req, res = createRequestResponse("DELETE", "/sample?:id=1", nil)
						handler(res, req)
					})
					It("returns 404 http status", func() {
						Expect(res.Code).To(Equal(404))
					})
					It("returns an error message in the response", func() {
						var response map[string]string
						err := json.Unmarshal(res.Body.Bytes(), &response)
						Expect(err).ToNot(HaveOccurred())
						Expect(response).To(HaveKey("error"))
					})
				})
			})
			Context("Given a repository with one item", func() {
				var idJoe string
				BeforeEach(func() {
					joe := aRecord("Joe", 30)
					idJoe, _ = repo.Save(ctx, &joe)
				})
				When("I call DELETE with an existing id", func() {
					BeforeEach(func() {
						req, res = createRequestResponse("DELETE", fmt.Sprintf("/sample?:id=%s", idJoe), nil)
						handler(res, req)
					})
					It("returns 200 http status", func() {
						Expect(res.Code).To(Equal(200))
					})
					It("deletes the record from the repository", func() {
						_, err := repo.Read(ctx, idJoe)
						Expect(err).To(MatchError(rest.ErrNotFound))
					})
				})
			})
			Context("When the repository returns an ErrPermissionDenied", func() {
				It("returns 403 http status", func() {
					repo.SetError(rest.ErrPermissionDenied)
					req, res = createRequestResponse("DELETE", "/sample?:id=1", nil)
					handler(res, req)

					Expect(res.Code).To(Equal(403))
				})
			})
			Context("When the repository returns an error", func() {
				BeforeEach(func() {
					repo.SetError(errors.New("unknown error"))
					req, res = createRequestResponse("DELETE", "/sample?:id=1", nil)
					handler(res, req)
				})
				It("returns 500 http status", func() {
					Expect(res.Code).To(Equal(500))
				})
				It("returns an error message in the response", func() {
					var response map[string]string
					err := json.Unmarshal(res.Body.Bytes(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(HaveKey("error"))
				})
			})
		})
	})
})

//func TestController_Put(t *testing.T) {
//	Convey("Given a read-only repository", t, func() {
//		handler, _ := createReadOnlyHandler(rest.Delete)
//		Convey("When I call Put id=1", func() {
//			req, res := createRequestResponse("PUT", "/sample?:id=1", nil)
//			handler(res, req)
//
//			Convey("It returns 405 http status", func() {
//				So(res.Code, ShouldEqual, 405)
//			})
//		})
//	})
//
//	Convey("Given an empty repository", t, func() {
//		handler, repo := createPersistableHandler(rest.Put)
//
//		Convey("When I call Put with an invalid request", func() {
//			req, res := createRequestResponse("PUT", "/sample?:id=1", nil)
//			handler(res, req)
//
//			Convey("It returns 422 http status", func() {
//				So(res.Code, ShouldEqual, 422)
//			})
//
//			Convey("It returns an error message in the response", func() {
//				var response map[string]string
//				if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
//					panic(err)
//				}
//				So(response, ShouldContainKey, "error")
//			})
//
//			Convey("It passes down the context", func() {
//				So(repo.Context.Value("test_key"), ShouldEqual, "test_value")
//			})
//		})
//
//		Convey("When I call Put id=1", func() {
//			req, res := createRequestResponse("PUT", "/sample?:id=1", aRecordBody("1", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 404 http status", func() {
//				So(res.Code, ShouldEqual, 404)
//			})
//
//			Convey("It returns an error message in the response", func() {
//				var response map[string]any
//				if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
//					panic(err)
//				}
//				So(response, ShouldContainKey, "error")
//			})
//		})
//
//		Convey("When an item is added", func() {
//			joe := aRecord("Joe", 30)
//			id, _ := repo.Save(&joe)
//
//			Convey("And I call Put", func() {
//				req, res := createRequestResponse("PUT", "/sample?:id="+id, aRecordBody(id, "John", 31))
//				handler(res, req)
//
//				Convey("It returns 200 http status", func() {
//					So(res.Code, ShouldEqual, 200)
//				})
//
//				Convey("It returns all data from the record", func() {
//					var response examples.SampleModel
//					if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
//						panic(err)
//					}
//					So(response.ID, ShouldEqual, id)
//					So(response.Name, ShouldEqual, "John")
//					So(response.Age, ShouldEqual, 31)
//				})
//			})
//		})
//
//		Convey("When the repository returns a ErrPermissionDenied", func() {
//			repo.SetError(rest.ErrPermissionDenied)
//
//			req, res := createRequestResponse("PUT", "/sample", aRecordBody("1", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 403 http status", func() {
//				So(res.Code, ShouldEqual, 403)
//			})
//		})
//
//		Convey("When the repository returns a ValidationError", func() {
//			repo.Error = &rest.ValidationError{Errors: map[string]string{
//				"field1": "not_valid",
//			}}
//
//			req, res := createRequestResponse("PUT", "/sample", aRecordBody("1", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 400 http status", func() {
//				So(res.Code, ShouldEqual, 400)
//			})
//
//			Convey("It returns a list of errors in the body", func() {
//				var parsed map[string]map[string]string
//				_ = json.Unmarshal(res.Body.Bytes(), &parsed)
//				So(parsed, ShouldContainKey, "errors")
//				So(parsed["errors"], ShouldContainKey, "field1")
//				So(parsed["errors"]["field1"], ShouldEqual, "not_valid")
//			})
//		})
//
//		Convey("When the repository returns an error", func() {
//			repo.Error = errors.New("unknown error")
//
//			req, res := createRequestResponse("PUT", "/sample", aRecordBody("1", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 500 http status", func() {
//				So(res.Code, ShouldEqual, 500)
//			})
//
//			Convey("It returns an error message in the response", func() {
//				var response map[string]string
//				if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
//					panic(err)
//				}
//				So(response, ShouldContainKey, "error")
//			})
//		})
//	})
//}
//
//func TestController_Post(t *testing.T) {
//	Convey("Given a read-only repository", t, func() {
//		handler, _ := createReadOnlyHandler(rest.Delete)
//		Convey("When I send valid data", func() {
//			req, res := createRequestResponse("POST", "/sample", aRecordBody("0", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 405 http status", func() {
//				So(res.Code, ShouldEqual, 405)
//			})
//		})
//	})
//
//	Convey("Given an empty repository", t, func() {
//		handler, repo := createPersistableHandler(rest.Post)
//
//		Convey("When I send valid data", func() {
//			req, res := createRequestResponse("POST", "/sample", aRecordBody("0", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 200 http status", func() {
//				So(res.Code, ShouldEqual, 200)
//			})
//
//			Convey("It adds the data to the repo", func() {
//				count, _ := repo.Count()
//				So(count, ShouldEqual, 1)
//			})
//
//			Convey("It returns the new id in the response", func() {
//				var response map[string]string
//				if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
//					panic(err)
//				}
//				So(response, ShouldContainKey, "id")
//				id := response["id"]
//				r, _ := repo.Read(id)
//				So(r.(examples.SampleModel).Name, ShouldEqual, "John Doe")
//			})
//
//			Convey("It passes down the context", func() {
//				So(repo.Context.Value("test_key"), ShouldEqual, "test_value")
//			})
//		})
//
//		Convey("When I send invalid data", func() {
//			req, res := createRequestResponse("POST", "/sample", strings.NewReader("BAD DATA"))
//			handler(res, req)
//
//			Convey("It returns 422 http status", func() {
//				So(res.Code, ShouldEqual, 422)
//			})
//
//			Convey("It does not adds any data to the repo", func() {
//				count, _ := repo.Count()
//				So(count, ShouldEqual, 0)
//			})
//		})
//
//		Convey("When the repository returns a ErrPermissionDenied", func() {
//			repo.SetError(rest.ErrPermissionDenied)
//
//			req, res := createRequestResponse("POST", "/sample", aRecordBody("0", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 403 http status", func() {
//				So(res.Code, ShouldEqual, 403)
//			})
//		})
//
//		Convey("When the repository returns a ValidationError", func() {
//			repo.SetError(&rest.ValidationError{Errors: map[string]string{
//				"field1": "not_valid",
//			}})
//
//			req, res := createRequestResponse("POST", "/sample", aRecordBody("0", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 400 http status", func() {
//				So(res.Code, ShouldEqual, 400)
//			})
//
//			Convey("It returns a list of errors in the body", func() {
//				var parsed map[string]map[string]string
//				_ = json.Unmarshal(res.Body.Bytes(), &parsed)
//				So(parsed, ShouldContainKey, "errors")
//				So(parsed["errors"], ShouldContainKey, "field1")
//				So(parsed["errors"]["field1"], ShouldEqual, "not_valid")
//			})
//		})
//
//		Convey("When the repository returns an error", func() {
//			repo.SetError(errors.New("unknown error"))
//
//			req, res := createRequestResponse("POST", "/sample", aRecordBody("0", "John Doe", 33))
//			handler(res, req)
//
//			Convey("It returns 500 http status", func() {
//				So(res.Code, ShouldEqual, 500)
//			})
//
//			Convey("It returns an error message in the response", func() {
//				var response map[string]string
//				if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
//					panic(err)
//				}
//				So(response, ShouldContainKey, "error")
//			})
//		})
//	})
//}

func aRecord(name string, age int) examples.SampleModel {
	return examples.SampleModel{Name: name, Age: age}
}

func aRecordBody(id string, name string, age int) io.Reader {
	r := aRecord(name, age)
	r.ID = id
	buf, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(buf)
}
