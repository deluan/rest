package rest

import (
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseOptions", func() {
	var c *Controller[any]

	BeforeEach(func() {
		c = &Controller[any]{}
	})

	Describe("Given no params", func() {
		It("returns an empty QueryOptions struct", func() {
			options, _ := c.parseOptions(url.Values{})
			Expect(options.Sort).To(BeEmpty())
			Expect(options.Order).To(BeEmpty())
			Expect(options.Offset).To(Equal(0))
			Expect(options.Max).To(Equal(0))
			Expect(options.Filters).To(BeEmpty())
		})
	})

	Describe("Given pagination params", func() {
		It("returns a properly filled QueryOptions struct", func() {
			params := url.Values{"_start": []string{"10"}, "_end": []string{"30"}, "_sort": []string{"name"}, "_order": []string{"DESC"}}
			options, _ := c.parseOptions(params)
			Expect(options.Sort).To(Equal("name"))
			Expect(options.Order).To(Equal("desc"))
			Expect(options.Offset).To(Equal(10))
			Expect(options.Max).To(Equal(20))
		})
	})

	Describe("Given individual filter params", func() {
		It("returns a properly filled QueryOptions struct", func() {
			params := url.Values{"name": []string{"joe"}, "age": []string{"30"}}
			options, _ := c.parseOptions(params)
			Expect(options.Filters).To(HaveLen(2))
			Expect(options.Filters["name"]).To(Equal("joe"))
			Expect(options.Filters["age"]).To(Equal("30"))
		})
	})

	Describe("Given duplicated individual filter params", func() {
		It("returns a properly filled QueryOptions struct", func() {
			params := url.Values{"name": []string{"joe", "cecilia"}}
			options, _ := c.parseOptions(params)
			Expect(options.Filters).To(HaveLen(1))
			Expect(options.Filters["name"]).To(ConsistOf("joe", "cecilia"))
		})
	})

	Describe("Given single filter param", func() {
		It("returns a properly filled QueryOptions struct", func() {
			params := url.Values{"_filters": []string{`{"name":"cecilia","age":"22"}`}}
			options, _ := c.parseOptions(params)
			Expect(options.Filters).To(HaveLen(2))
			Expect(options.Filters["name"]).To(Equal("cecilia"))
			Expect(options.Filters["age"]).To(Equal("22"))
		})
	})

	Describe("Given an invalid single filter param", func() {
		It("ignores the filter", func() {
			params := url.Values{"_filters": []string{`{"name":"cecilia","age":MISSING_QUOTES}`}}
			options, _ := c.parseOptions(params)
			Expect(options.Filters).To(HaveLen(0))
		})
	})
})
