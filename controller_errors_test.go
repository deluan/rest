package rest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/deluan/rest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestController_WrappedErrors(t *testing.T) {
	wrap := func(err error) error { return fmt.Errorf("wrapped: %w", err) }
	validationErr := &rest.ValidationError{Errors: map[string]string{"field1": "not_valid"}}

	cases := []struct {
		name    string
		handler handlerWrapper
		method  string
		err     error
		status  int
	}{
		{"Get with wrapped ErrNotFound", rest.Get, "GET", wrap(rest.ErrNotFound), 404},
		{"Get with wrapped ErrPermissionDenied", rest.Get, "GET", wrap(rest.ErrPermissionDenied), 403},
		{"GetAll with wrapped ErrPermissionDenied", rest.GetAll, "GET", wrap(rest.ErrPermissionDenied), 403},
		{"Put with wrapped ErrNotFound", rest.Put, "PUT", wrap(rest.ErrNotFound), 404},
		{"Put with wrapped ErrPermissionDenied", rest.Put, "PUT", wrap(rest.ErrPermissionDenied), 403},
		{"Put with wrapped ValidationError", rest.Put, "PUT", wrap(validationErr), 400},
		{"Post with wrapped ErrPermissionDenied", rest.Post, "POST", wrap(rest.ErrPermissionDenied), 403},
		{"Post with wrapped ValidationError", rest.Post, "POST", wrap(validationErr), 400},
		{"Delete with wrapped ErrNotFound", rest.Delete, "DELETE", wrap(rest.ErrNotFound), 404},
		{"Delete with wrapped ErrPermissionDenied", rest.Delete, "DELETE", wrap(rest.ErrPermissionDenied), 403},
	}

	for _, tc := range cases {
		Convey("When the repository returns a "+tc.name, t, func() {
			handler, repo := createPersistableHandler(tc.handler)
			repo.Error = tc.err

			req, res := createRequestResponse(tc.method, "/sample?:id=1", strings.NewReader(`{"name":"John Doe","age":33}`))
			handler(res, req)

			Convey(fmt.Sprintf("It returns %d http status", tc.status), func() {
				So(res.Code, ShouldEqual, tc.status)
			})
		})
	}
}
