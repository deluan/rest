package rest

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http/httptest"
)

type payload struct{ Field1 int }

var _ = Describe("Response renderers", func() {
	var recorder *httptest.ResponseRecorder
	BeforeEach(func() {
		recorder = httptest.NewRecorder()
	})

	Describe("RespondWithJSON", func() {
		var response any

		Context("Given a success payload", func() {
			BeforeEach(func() {
				response = payload{123}
				_ = RespondWithJSON(recorder, 200, response)
			})

			It("sets the right content-type", func() {
				Expect(recorder.Header()["Content-Type"]).To(ContainElement("application/json"))
			})

			It("sends the correct status", func() {
				Expect(recorder.Code).To(Equal(200))
			})

			It("sends the payload", func() {
				actual := &payload{}
				err := json.Unmarshal(recorder.Body.Bytes(), actual)
				Expect(err).NotTo(HaveOccurred())
				Expect(*actual).To(Equal(response))
			})
		})

		Context("Given an invalid payload", func() {
			BeforeEach(func() {
				response = func() {}
			})

			It("returns an error", func() {
				err := RespondWithJSON(recorder, 200, response)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("RespondWithError", func() {
		Context("Given an error payload", func() {
			BeforeEach(func() {
				_ = RespondWithError(recorder, 400, "error message")
			})

			It("sets the right content-type", func() {
				Expect(recorder.Header()["Content-Type"]).To(ContainElement("application/json"))
			})

			It("sends the correct status", func() {
				Expect(recorder.Code).To(Equal(400))
			})

			It("sends the payload", func() {
				Expect(recorder.Body.String()).To(Equal(`{"error":"error message"}`))
			})
		})
	})
})
