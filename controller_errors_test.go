package rest_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/deluan/rest"
	"github.com/deluan/rest/examples"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Wrapped repository errors", func() {
	type handlerFunc func(rest.Repository[examples.SampleModel]) http.HandlerFunc
	wrap := func(err error) error { return fmt.Errorf("wrapped: %w", err) }
	validationErr := &rest.ValidationError{Errors: map[string]string{"field1": "not_valid"}}

	DescribeTable("maps wrapped errors to the same http status as bare errors",
		func(newHandler handlerFunc, method string, err error, status int) {
			repo := examples.NewPersistableSampleRepository()
			repo.SetError(err)
			handler := newHandler(repo)

			req, res := createRequestResponse(method, "/sample?:id=1", strings.NewReader(`{"name":"John Doe","age":33}`))
			handler(res, req)

			Expect(res.Code).To(Equal(status))
		},
		Entry("Get with ErrNotFound", handlerFunc(rest.Get[examples.SampleModel]), "GET", wrap(rest.ErrNotFound), 404),
		Entry("Get with ErrPermissionDenied", handlerFunc(rest.Get[examples.SampleModel]), "GET", wrap(rest.ErrPermissionDenied), 403),
		Entry("GetAll with ErrPermissionDenied", handlerFunc(rest.GetAll[examples.SampleModel]), "GET", wrap(rest.ErrPermissionDenied), 403),
		Entry("Put with ErrNotFound", handlerFunc(rest.Put[examples.SampleModel]), "PUT", wrap(rest.ErrNotFound), 404),
		Entry("Put with ErrPermissionDenied", handlerFunc(rest.Put[examples.SampleModel]), "PUT", wrap(rest.ErrPermissionDenied), 403),
		Entry("Put with ValidationError", handlerFunc(rest.Put[examples.SampleModel]), "PUT", wrap(validationErr), 400),
		Entry("Post with ErrPermissionDenied", handlerFunc(rest.Post[examples.SampleModel]), "POST", wrap(rest.ErrPermissionDenied), 403),
		Entry("Post with ValidationError", handlerFunc(rest.Post[examples.SampleModel]), "POST", wrap(validationErr), 400),
		Entry("Delete with ErrNotFound", handlerFunc(rest.Delete[examples.SampleModel]), "DELETE", wrap(rest.ErrNotFound), 404),
		Entry("Delete with ErrPermissionDenied", handlerFunc(rest.Delete[examples.SampleModel]), "DELETE", wrap(rest.ErrPermissionDenied), 403),
	)
})

type box[V any] struct{ Value V }

type boxRepo[V any] struct{}

func (boxRepo[V]) Count(context.Context, ...rest.QueryOptions) (int64, error) { return 0, nil }
func (boxRepo[V]) Read(context.Context, string) (*box[V], error)              { return nil, rest.ErrNotFound }
func (boxRepo[V]) ReadAll(context.Context, ...rest.QueryOptions) ([]box[V], error) {
	return nil, nil
}

var _ = Describe("Entity name in error messages", func() {
	It("uses the bare type name, without the package", func() {
		handler := rest.Get(rest.Repository[examples.SampleModel](examples.NewPersistableSampleRepository()))
		req, res := createRequestResponse("GET", "/sample?:id=1", nil)
		handler(res, req)

		Expect(res.Code).To(Equal(404))
		Expect(res.Body.String()).To(MatchJSON(`{"error":"SampleModel(id:1) not found"}`))
	})

	It("drops the type arguments of generic types", func() {
		handler := rest.Get(rest.Repository[box[examples.SampleModel]](boxRepo[examples.SampleModel]{}))
		req, res := createRequestResponse("GET", "/box?:id=1", nil)
		handler(res, req)

		Expect(res.Code).To(Equal(404))
		Expect(res.Body.String()).To(MatchJSON(`{"error":"box(id:1) not found"}`))
	})
})
