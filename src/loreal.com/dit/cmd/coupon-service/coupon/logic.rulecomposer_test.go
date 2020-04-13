package coupon

import (
	// "encoding/json"
	// "fmt"
	// "reflect"
	"testing"
	"time"

	"loreal.com/dit/cmd/coupon-service/base"

	. "github.com/chenhg5/collection"
	"github.com/mitchellh/mapstructure"
	. "github.com/smartystreets/goconvey/convey"
	// "bou.ke/monkey"
)

func Test_RedeemTimesBodyComposer_compose(t *testing.T) {
	Convey("Given a RedeemTimesBodyComposer instance and some input", t, func() {
		var composer RedeemTimesBodyComposer
		// var ruleInternalID = base.RandString(4)
		ruleBody := map[string]interface{}{
			"times": 123,
		}
		Convey("Call Compose", func() {
			rrf, _ := composer.Compose(nil, "", ruleBody)
			Convey("The composed value should contain correct value", func() {

				So(Collect(rrf).Has("times"), ShouldBeTrue)
			})
		})
	})

	// var st = time.Now().Local().AddDate(0, 0, 0)
	// fmt.Print(st.String())

	// t.Errorf(st.Format("2006-01-02 15:04:05 -07"))

}

func Test_RedeemInCurrentNatureTimeUnitBodyComposer_compose(t *testing.T) {
	Convey("Given a RedeemInCurrentNatureTimeUnitBodyComposer instance and a rule body with positive value", t, func() {
		var composer RedeemInCurrentNatureTimeUnitBodyComposer
		// var ruleInternalID = base.RandString(4)
		ruleBody := map[string]interface{}{
			"unit":         "MONTH",
			"endInAdvance": 10,
		}
		Convey("Call Compose", func() {
			rrf, _ := composer.Compose(nil, "", ruleBody)
			Convey("The composed value should contain correct value", func() {
				So(Collect(rrf).Has("endInAdvance"), ShouldBeTrue)
				So(Collect(rrf).Has("unit"), ShouldBeTrue)
				var ntu natureTimeUnit
				mapstructure.Decode(rrf, &ntu)
				So(ntu.Unit, ShouldEqual, "MONTH")
				So(ntu.EndInAdvance, ShouldEqual, 10)
			})
		})
	})

	Convey("Given a RedeemInCurrentNatureTimeUnitBodyComposer instance and a rule body with negative value", t, func() {
		var composer RedeemInCurrentNatureTimeUnitBodyComposer
		// var ruleInternalID = base.RandString(4)
		ruleBody := map[string]interface{}{
			"unit":         "MONTH",
			"endInAdvance": -10,
		}
		Convey("Call Compose", func() {
			rrf, _ := composer.Compose(nil, "", ruleBody)
			Convey("The composed value should contain correct value", func() {
				So(Collect(rrf).Has("endInAdvance"), ShouldBeTrue)
				So(Collect(rrf).Has("unit"), ShouldBeTrue)
				var ntu natureTimeUnit
				mapstructure.Decode(rrf, &ntu)
				So(ntu.Unit, ShouldEqual, "MONTH")
				So(ntu.EndInAdvance, ShouldEqual, -10)
			})
		})
	})

}

func Test_RedeemBySameBrandBodyComposer_compose(t *testing.T) {
	var composer RedeemBySameBrandBodyComposer
	// var ruleInternalID = base.RandString(4)
	ruleBody := map[string]interface{}{}
	Convey("Given a RedeemBySameBrandBodyComposer instance and some input", t, func() {
		var brand = base.RandString(4)
		var requester = _aRequester("", nil, brand)
		// monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
		// 	return true
		// })
		Convey("Call Compose", func() {

			rrf, _ := composer.Compose(requester, "", ruleBody)
			Convey("The composed value should contain brand info", func() {
				// So(ruleInternalID, ShouldEqual, rrf.RuleInternalID)
				// var sb sameBrand
				// _ = json.Unmarshal([]byte(rrf.RuleBody), &sb)
				So(Collect(rrf).Has("brand"), ShouldBeTrue)
				So(rrf["brand"], ShouldEqual, brand)
			})
		})
	})

	Convey("Given a requester with no brand", t, func() {
		var requester = _aRequester("", nil, "")
		Convey("Call Compose", func() {
			_, err := composer.Compose(requester, "", ruleBody)
			Convey("The call should failed with ErrRequesterHasNoBrand", func() {
				So(err, ShouldEqual, &ErrRequesterHasNoBrand)
			})
		})
	})
}

func Test_ValidPeriodWithOffsetBodyComposer_compose(t *testing.T) {
	var composer RedeemPeriodWithOffsetBodyComposer
	// var ruleInternalID = base.RandString(4)
	Convey("Given a RedeemPeriodWithOffsetBodyComposer instance and some input", t, func() {
		var offSetFromAppliedDay = r.Intn(1000)
		var span = r.Intn(1000)
		ruleBody := map[string]interface{}{
			"offSetFromAppliedDay": offSetFromAppliedDay,
			"timeSpan":             span,
		}
		// var bodyString = fmt.Sprintf(`{"offSetFromAppliedDay": %d,"timeSpan": %d}`, offSetFromAppliedDay, span)
		// monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
		// 	return true
		// })
		Convey("Call Compose", func() {
			rrf, _ := composer.Compose(nil, "", ruleBody)
			Convey("The composed value should contain an offset value", func() {
				// So(ruleInternalID, ShouldEqual, rrf.RuleInternalID)

				var st = time.Now().Local().AddDate(0, 0, offSetFromAppliedDay)
				var end = st.AddDate(0, 0, span)
				// var ts timeSpan
				// _ = json.Unmarshal([]byte(rrf.RuleBody), &ts)
				So(Collect(rrf).Has("startTime"), ShouldBeTrue)
				So(Collect(rrf).Has("endTime"), ShouldBeTrue)
				So(rrf["startTime"], ShouldEqual, st.Format("2006-01-02 15:04:05 -07"))
				So(rrf["endTime"], ShouldEqual, end.Format("2006-01-02 15:04:05 -07"))
			})
		})
	})

}
