package base

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type errorBody struct {
	Code    	int    		`json:"error-code"`
	Msg        	string    	`json:"error-message"`
}

const sampleCode 	int = 100
const sampleMsg   	string = "Hello world"

func TestBaseerror(t *testing.T) {
	Convey("Given an ErrorWithCode object", t, func() {
		var e = ErrorWithCode {
			Code : sampleCode,
			Message : sampleMsg,
		}
		Convey("The error message Unmarshal by json", func() {
			var js = e.Error()
			var eb errorBody
			err := json.Unmarshal([]byte(js), &eb)
			Convey("The value should not be changed", func() {
				So(err, ShouldBeNil)
				So(eb.Code, ShouldEqual, sampleCode )
				So(eb.Msg, ShouldEqual, sampleMsg )
			})
		})
	})
}