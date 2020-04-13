package base

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsBlankString(t *testing.T) {
	Convey("Given a hello world string", t, func() {
		str :=  "Hello world"
		Convey("The str should not be blank", func() {
			So(IsBlankString(str), ShouldBeFalse)
		})
	})

	Convey("Given a string with nothing", t, func() {
		str :=  ""
		Convey("The str should be blank", func() {
			So(IsBlankString(str), ShouldBeTrue)
		})
	})

	Convey("Given a string with some blank characters", t, func() {
		str :=  "    "
		Convey("The str should be blank", func() {
			So(IsBlankString(str), ShouldBeTrue)
		})
	})
}

func TestIsEmptyString(t *testing.T) {
	Convey("Given a hello world string", t, func() {
		str :=  "Hello world"
		Convey("The str should not be empty", func() {
			So(IsEmptyString(str), ShouldBeFalse)
		})
	})

	Convey("Given a string with nothing", t, func() {
		str :=  ""
		Convey("The str should be empty", func() {
			So(IsEmptyString(str), ShouldBeTrue)
		})
	})

	Convey("Given a string with some blank characters", t, func() {
		str :=  "    "
		Convey("The str should not be empty", func() {
			So(IsEmptyString(str), ShouldBeFalse)
		})
	})
}

func TestIsValidUUID(t *testing.T) {
	Convey("Given an UUID string", t, func() {
		str :=  "66e382c4-b859-4e46-9a88-d875fbdaf366"
		Convey("The str should be UUID", func() {
			So(IsValidUUID(str), ShouldBeTrue)
		})
	})

	Convey("Given a hello world string", t, func() {
		str :=  "Hello world"
		Convey("The str should not be UUID", func() {
			So(IsValidUUID(str), ShouldBeFalse)
		})
	})
}
