package coupon

import (
	// "encoding/json"
	"fmt"
	// "reflect"
	"testing"
	"time"

	// "loreal.com/dit/cmd/coupon-service/base"

	"bou.ke/monkey"
	"github.com/jinzhu/now"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_ValidPeriodJudge_judgeCoupon(t *testing.T) {
	var judge RedeemPeriodJudge
	Convey("Given a regular time span and include today", t, func() {
		ruleBody := map[string]interface{}{
			"startTime": "2020-01-01 00:00:00 +08",
			"endTime":   "3020-02-14 00:00:00 +08",
		}
		// var ruleBody = `{"startTime": "2020-01-01 00:00:00 +08",  "endTime": "3020-02-14 00:00:00 +08" }`
		//base.RandString(4)
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(nil, "", ruleBody, nil)
			Convey("Should no errors", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a past time span", t, func() {
		ruleBody := map[string]interface{}{
			"startTime": "1989-06-04 01:23:45 +08",
			"endTime":   "2008-02-08 20:08:00 +08",
		}
		// var ruleBody = `{"startTime": "1989-06-04 01:23:45 +08",  "endTime": "2008-02-08 20:08:00 +08" }`
		//base.RandString(4)
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(nil, "", ruleBody, nil)
			Convey("Should ErrCouponRulesRedemptionExpired", func() {
				So(err, ShouldEqual, &ErrCouponRulesRedemptionExpired)
			})
		})
	})

	Convey("Given a future time span", t, func() {
		ruleBody := map[string]interface{}{
			"startTime": "3456-07-08 09:10:12 +08",
			"endTime":   "3456-07-08 09:20:12 +08",
		}
		// var ruleBody = `{"startTime": "3456-07-08 09:10:12 +08",  "endTime": "3456-07-08 09:20:12 +08" }`
		//base.RandString(4)
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(nil, "", ruleBody, nil)
			Convey("Should ErrCouponRulesRedemptionNotStart", func() {
				So(err, ShouldEqual, &ErrCouponRulesRedemptionNotStart)
			})
		})
	})

	Convey("Given some time spans with bad format", t, func() {
		Convey("bad start format", func() {
			ruleBody := map[string]interface{}{
				"startTime": 1234567890,
				"endTime":   "3456-07-08 09:20:12 +08",
			}
			// var ruleBody = `{"startTime": 1234567890,  "endTime": "3456-07-08 09:20:12 +08" }`
			err := judge.JudgeCoupon(nil, "", ruleBody, nil)
			Convey("Should ErrCouponRulesBadFormat", func() {
				So(err, ShouldEqual, &ErrCouponRulesBadFormat)
			})
		})

		Convey("bad end format", func() {
			ruleBody := map[string]interface{}{
				"startTime": "3456-07-08 09:10:12 +08",
				"endTime":   1234567890,
			}
			// var ruleBody = `{"startTime": "3456-07-08 09:10:12 +08",  "endTime": "3456=07-08 09:20:12 +08" }`
			err := judge.JudgeCoupon(nil, "", ruleBody, nil)
			Convey("Should ErrCouponRulesBadFormat", func() {
				So(err, ShouldEqual, &ErrCouponRulesBadFormat)
			})
		})
	})
}

func Test_RedeemInCurrentNatureMonthSeasonYear_judgeCoupon(t *testing.T) {
	var judge RedeemInCurrentNatureTimeUnitJudge
	ruleBody := map[string]interface{}{
		"unit":         "MONTH",
		"endInAdvance": 0,
	}
	Convey("Given a regular sample", t, func() {
		// var ruleBody = `{"startTime": "2020-01-01 00:00:00 +08",  "endTime": "3020-02-14 00:00:00 +08" }`
		//base.RandString(4)
		c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(nil, "", ruleBody, c)
			Convey("Should no errors", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a coupon applied last month", t, func() {
		c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
		ct := c.CreatedTime.AddDate(0, 0, -31)
		c.CreatedTime = &ct
		fmt.Print(c.CreatedTime)
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(nil, "", ruleBody, c)
			Convey("Should ErrCouponRulesRedemptionExpired", func() {
				So(err, ShouldEqual, &ErrCouponRulesRedemptionExpired)
			})
		})
	})

	Convey("Given a coupon applied before the month, maybe caused by daylight saving time", t, func() {
		c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
		ct := c.CreatedTime.AddDate(0, 0, 31)
		c.CreatedTime = &ct
		// var ruleBody = `{"startTime": "3456-07-08 09:10:12 +08",  "endTime": "3456-07-08 09:20:12 +08" }`
		//base.RandString(4)
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(nil, "", ruleBody, c)
			Convey("Should ErrCouponRulesRedemptionNotStart", func() {
				So(err, ShouldEqual, &ErrCouponRulesRedemptionNotStart)
			})
		})
	})

	Convey("Given a time unit with unsupport unit", t, func() {
		Convey("season unit...", func() {
			rb := map[string]interface{}{
				"unit":         "SEASON",
				"endInAdvance": 0,
			}
			err := judge.JudgeCoupon(nil, "", rb, nil)
			Convey("Should ErrCouponRulesUnsuportTimeUnit", func() {
				So(err, ShouldEqual, &ErrCouponRulesUnsuportTimeUnit)
			})
		})

		Convey("year unit...", func() {
			rb := map[string]interface{}{
				"unit":         "YEAR",
				"endInAdvance": 0,
			}
			err := judge.JudgeCoupon(nil, "", rb, nil)
			Convey("Should ErrCouponRulesUnsuportTimeUnit", func() {
				So(err, ShouldEqual, &ErrCouponRulesUnsuportTimeUnit)
			})
		})
	})

	Convey("Given a sample with not default endInAdvance", t, func() {
		rb := map[string]interface{}{
			"unit":         "MONTH",
			"endInAdvance": 10,
		}
		c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
		nov := now.With(time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC))
		pg1 := monkey.Patch(now.With, func(time.Time) *now.Now {
			return nov
		})

		Convey("assume in valid period", func() {
			ct := time.Date(2020, time.November, 15, 0, 0, 0, 0, time.UTC)
			pg2 := monkey.Patch(time.Now, func() time.Time {
				return ct
			})
			Convey("Call JudgeCoupon", func() {
				err := judge.JudgeCoupon(nil, "", rb, c)
				Convey("Should no errors", func() {
					So(err, ShouldBeNil)
				})
			})
			pg2.Unpatch()
		})

		Convey("assume out of period", func() {
			ct := time.Date(2020, time.November, 30, 0, 0, 0, 0, time.UTC)
			pg2 := monkey.Patch(time.Now, func() time.Time {
				return ct
			})
			Convey("Call JudgeCoupon", func() {
				err := judge.JudgeCoupon(nil, "", rb, c)
				Convey("Should ErrCouponRulesRedemptionExpired", func() {
					So(err, ShouldEqual, &ErrCouponRulesRedemptionExpired)
				})
			})
			pg2.Unpatch()
		})

		pg1.Unpatch()
	})

	Convey("Given a sample with not default endInAdvance", t, func() {
		rb := map[string]interface{}{
			"unit":         "MONTH",
			"endInAdvance": -10,
		}
		c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
		nov := now.With(time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC))
		pg1 := monkey.Patch(now.With, func(time.Time) *now.Now {
			return nov
		})

		Convey("assume in valid period", func() {
			ct := time.Date(2020, time.December, 5, 0, 0, 0, 0, time.UTC)
			pg2 := monkey.Patch(time.Now, func() time.Time {
				return ct
			})
			Convey("Call JudgeCoupon", func() {
				err := judge.JudgeCoupon(nil, "", rb, c)
				Convey("Should no errors", func() {
					So(err, ShouldBeNil)
				})
			})
			pg2.Unpatch()
		})

		Convey("assume out of period", func() {
			ct := time.Date(2020, time.December, 15, 0, 0, 0, 0, time.UTC)
			pg2 := monkey.Patch(time.Now, func() time.Time {
				return ct
			})
			Convey("Call JudgeCoupon", func() {
				err := judge.JudgeCoupon(nil, "", rb, c)
				Convey("Should ErrCouponRulesRedemptionExpired", func() {
					So(err, ShouldEqual, &ErrCouponRulesRedemptionExpired)
				})
			})
			pg2.Unpatch()
		})

		pg1.Unpatch()
	})
}

func Test_RedeemTimesJudge_judgeCoupon(t *testing.T) {
	var judge RedeemTimesJudge
	var c Coupon
	Convey("Given a valid redeem times", t, func() {
		ruleBody := map[string]interface{}{
			"times": 2,
		}
		// var ruleBody = `{"times": 2}`
		Convey("Fisrt give some redeem logs which less than the coupon max redeem times", func() {
			monkey.Patch(getCouponTransactionsWithType, func(_ string, _ TransType) ([]*Transaction, error) {
				return make([]*Transaction, 1, 1), nil
			})
			Convey("Call JudgeCoupon", func() {
				err := judge.JudgeCoupon(nil, "", ruleBody, &c)
				Convey("Should no errors", func() {
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("Second give no redeem logs", func() {
			monkey.Patch(getCouponTransactionsWithType, func(_ string, _ TransType) ([]*Transaction, error) {
				return nil, nil
			})
			Convey("Call JudgeCoupon", func() {
				err := judge.JudgeCoupon(nil, "", ruleBody, &c)
				Convey("Should no errors", func() {
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("Third give some redeem logs which greater than the coupon max redeem times", func() {
			monkey.Patch(getCouponTransactionsWithType, func(_ string, _ TransType) ([]*Transaction, error) {
				return make([]*Transaction, 3, 3), nil
			})
			Convey("Call JudgeCoupon", func() {
				err := judge.JudgeCoupon(nil, "", ruleBody, &c)
				Convey("Should has ErrCouponRulesRedeemTimesExceeded", func() {
					So(err, ShouldEqual, &ErrCouponRulesRedeemTimesExceeded)
				})
			})
		})

		Convey("If has some db err....", func() {
			monkey.Patch(getCouponTransactionsWithType, func(_ string, _ TransType) ([]*Transaction, error) {
				return nil, fmt.Errorf("hehehe")
			})
			Convey("Call JudgeCoupon", func() {
				err := judge.JudgeCoupon(nil, "", ruleBody, &c)
				Convey("Should no errors", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}

func Test_RedeemBySameBrandJudge_judgeCoupon(t *testing.T) {
	var judge RedeemBySameBrandJudge
	var brand = "Lancome"
	Convey("Given a reqeust with no brand", t, func() {
		var requester = _aRequester("", nil, "")
		ruleBody := map[string]interface{}{}
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(requester, "", ruleBody, nil)
			Convey("Should has ErrRequesterHasNoBrand", func() {
				So(err, ShouldEqual, &ErrRequesterHasNoBrand)
			})
		})
	})

	Convey("Given a bad fromat rule body", t, func() {
		var requester = _aRequester("", nil, brand)
		ruleBody := map[string]interface{}{
			"bra--nd": "Lancome",
		}
		// var ruleBody = `{"brand"="Lancome"}`
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(requester, "", ruleBody, nil)
			Convey("Should has error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a request with wrong brand", t, func() {
		var requester = _aRequester("", nil, brand)
		ruleBody := map[string]interface{}{
			"brand": "Lancome+bad+brand",
		}
		// var ruleBody = `{"brand":"Lancome+bad+brand"}`
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(requester, "", ruleBody, nil)
			Convey("Should has ErrRedeemWithDiffBrand", func() {
				So(err, ShouldEqual, &ErrRedeemWithDiffBrand)
			})
		})
	})

	Convey("Given a request with correct brand", t, func() {
		var requester = _aRequester("", nil, brand)
		ruleBody := map[string]interface{}{
			"brand": "Lancome",
		}

		// var ruleBody = `{"brand":"Lancome"}`
		Convey("Call JudgeCoupon", func() {
			err := judge.JudgeCoupon(requester, "", ruleBody, nil)
			Convey("Should has no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func Test_ApplyTimesJudge_judgeCoupon(t *testing.T) {
	var judge ApplyTimesJudge
	Convey("The data base has something wrong... ", t, func() {
		monkey.Patch(getCoupons, func(_ string, _ string) ([]*Coupon, error) {
			return nil, fmt.Errorf("hehehe")
		})
		Convey("Call JudgeTemplate", func() {
			ruleBody := map[string]interface{}{}
			err := judge.JudgeTemplate("", "", ruleBody, nil)
			Convey("Should has error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a coupon template which is expired", t, func() {
		ruleBody := map[string]interface{}{
			"inDays": 365,
			"times":  2,
		}
		// var ruleBody = `{"inDays": 365, "times": 2 }`
		var pct PublishedCouponType
		pct.CreatedTime = time.Now().Local().AddDate(0, 0, -366)
		monkey.Patch(getCoupons, func(_ string, _ string) ([]*Coupon, error) {
			return nil, nil
		})
		Convey("Call JudgeTemplate", func() {
			err := judge.JudgeTemplate("", "", ruleBody, &pct)
			Convey("Should has ErrCouponRulesApplyTimeExpired", func() {
				So(err, ShouldEqual, &ErrCouponRulesApplyTimeExpired)
			})
		})
	})

	Convey("Set the user had applyed coupons and reach the max times", t, func() {
		ruleBody := map[string]interface{}{
			"inDays": 365,
			"times":  2,
		}
		// var ruleBody = `{"inDays": 365, "times": 2 }`
		var pct PublishedCouponType
		pct.CreatedTime = time.Now().Local()
		monkey.Patch(getCoupons, func(_ string, _ string) ([]*Coupon, error) {
			return make([]*Coupon, 2, 2), nil
		})
		Convey("Call JudgeTemplate", func() {
			err := judge.JudgeTemplate("", "", ruleBody, &pct)
			Convey("Should has ErrCouponRulesApplyTimesExceeded", func() {
				So(err, ShouldEqual, &ErrCouponRulesApplyTimesExceeded)
			})
		})
	})

	Convey("Set the user has not reach the max appling times", t, func() {
		ruleBody := map[string]interface{}{
			"inDays": 365,
			"times":  2,
		}
		// var ruleBody = `{"inDays": 365, "times": 2 }`
		var pct PublishedCouponType
		pct.CreatedTime = time.Now().Local()
		monkey.Patch(getCoupons, func(_ string, _ string) ([]*Coupon, error) {
			return make([]*Coupon, 1, 2), nil
		})
		Convey("Call JudgeTemplate", func() {
			err := judge.JudgeTemplate("", "", ruleBody, &pct)
			Convey("Should has ErrCouponRulesApplyTimesExceeded", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
