package rest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type payload struct{ Field1 int }

func TestRespondWithJSON(t *testing.T) {
	Convey("Given a success payload", t, func() {
		response := payload{123}
		recorder := httptest.NewRecorder()
		RespondWithJSON(recorder, 200, response)

		Convey("It sets the right content-type", func() {
			So(recorder.HeaderMap["Content-Type"], ShouldContain, "application/json")
		})
		Convey("It sends the correct status", func() {
			So(recorder.Code, ShouldEqual, 200)
		})
		Convey("It sends the payload", func() {
			actual := &payload{}
			if err := json.Unmarshal([]byte(recorder.Body.String()), actual); err != nil {
				panic(err)
			}
			So(*actual, ShouldResemble, response)
		})
	})
	Convey("Given an invalid payload", t, func() {
		response := func() {}
		recorder := httptest.NewRecorder()
		err := RespondWithJSON(recorder, 200, response)
		Convey("It returns an error", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

func TestRespondWithError(t *testing.T) {
	Convey("Given an error payload", t, func() {
		recorder := httptest.NewRecorder()
		RespondWithError(recorder, 400, "error message")
		Convey("It sets the right content-type", func() {
			So(recorder.HeaderMap["Content-Type"], ShouldContain, "application/json")
		})
		Convey("It sends the correct status", func() {
			So(recorder.Code, ShouldEqual, 400)
		})
		Convey("It sends the payload", func() {
			So(recorder.Body.String(), ShouldEqual, `{"error":"error message"}`)
		})
	})
}
