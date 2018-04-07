package rest

import (
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

var logger = logrus.New()

func Test_parseOptions(t *testing.T) {
	c := &Controller{Logger: logger}

	Convey("Given no params", t, func() {
		options := c.parseOptions(url.Values{})
		Convey("It returns an empty QueryOptions struct", func() {
			So(options.Sort, ShouldBeEmpty)
			So(options.Order, ShouldBeEmpty)
			So(options.Offset, ShouldEqual, 0)
			So(options.Max, ShouldEqual, 0)
			So(options.Filters, ShouldBeEmpty)
		})
	})

	Convey("Given pagination params", t, func() {
		params := url.Values{"_start": []string{"10"}, "_end": []string{"30"}, "_sort": []string{"name"}, "_order": []string{"DESC"}}
		options := c.parseOptions(params)

		Convey("it  returns a proper filled QueryOptions struct", func() {
			So(options.Sort, ShouldEqual, "name")
			So(options.Order, ShouldEqual, "desc")
			So(options.Offset, ShouldEqual, 10)
			So(options.Max, ShouldEqual, 20)
		})
	})

	Convey("Given individual filter params", t, func() {
		params := url.Values{"name": []string{"joe"}, "age": []string{"30"}}
		options := c.parseOptions(params)

		Convey("it  returns a proper filled QueryOptions struct", func() {
			So(options.Filters, ShouldHaveLength, 2)
			So(options.Filters["name"], ShouldEqual, "joe")
			So(options.Filters["age"], ShouldEqual, "30")
		})
	})

	Convey("Given single filter param", t, func() {
		params := url.Values{"_filters": []string{`{"name":"cecilia","age":"22"}`}}
		options := c.parseOptions(params)

		Convey("it returns a proper filled QueryOptions struct", func() {
			So(options.Filters, ShouldHaveLength, 2)
			So(options.Filters["name"], ShouldEqual, "cecilia")
			So(options.Filters["age"], ShouldEqual, "22")
		})
	})

	Convey("Given an invalid single filter param", t, func() {
		params := url.Values{"_filters": []string{`{"name":"cecilia","age":MISSING_QUOTES}`}}
		options := c.parseOptions(params)

		Convey("it ignores the filter", func() {
			So(options.Filters, ShouldHaveLength, 0)
		})
	})
}

func init() {
	logger.SetLevel(logrus.FatalLevel)
}
