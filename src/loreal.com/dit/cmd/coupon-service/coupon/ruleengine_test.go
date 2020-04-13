package coupon

import (
	// "encoding/json"
	// "fmt"
	// "log"
	"reflect"
	"testing"
	"time"

	"loreal.com/dit/cmd/coupon-service/base"

	"bou.ke/monkey"
	. "github.com/smartystreets/goconvey/convey"
)

func _aREDEEM_TIMES_RuleRef(times int) (string, map[string]interface{}) {
	ruleBody := map[string]interface{}{
		"times": times,
	}
	return "REDEEM_TIMES", ruleBody
}

func _aVAILD_PERIOD_WITH_OFFSET_RuleRef(offset int, span int) (string, map[string]interface{}) {
	ruleBody := map[string]interface{}{
		"offSetFromAppliedDay": offset,
		"timeSpan":             span,
	}
	return "REDEEM_PERIOD_WITH_OFFSET", ruleBody
}

func _aAPPLY_TIMES_RuleRef(indays int, times int) (string, map[string]interface{}) {
	ruleBody := map[string]interface{}{
		"inDays": indays,
		"times":  times,
	}
	return "APPLY_TIMES", ruleBody
}

func _aREDEEM_BY_SAME_BRAND_RuleRef() (string, map[string]interface{}) {
	ruleBody := map[string]interface{}{}
	return "REDEEM_BY_SAME_BRAND", ruleBody
}
func _someRules() map[string]map[string]interface{} {
	var rrs = make(map[string]map[string]interface{}, 4)
	rid, rbd := _aREDEEM_TIMES_RuleRef(1)
	rrs[rid] = rbd
	rid, rbd = _aVAILD_PERIOD_WITH_OFFSET_RuleRef(0, 100)
	rrs[rid] = rbd
	rid, rbd = _aAPPLY_TIMES_RuleRef(0, 1)
	rrs[rid] = rbd
	rid, rbd = _aREDEEM_BY_SAME_BRAND_RuleRef()
	rrs[rid] = rbd
	return rrs
}

func _aPublishedCouponType(rules map[string]map[string]interface{}) *PublishedCouponType {
	var pct = PublishedCouponType{
		ID:                  base.RandString(4),
		Name:                base.RandString(4),
		TemplateID:          base.RandString(4),
		Description:         base.RandString(4),
		InternalDescription: base.RandString(4),
		State:               0,
		Publisher:           base.RandString(4),
		VisibleStartTime:    time.Now().Local().AddDate(0, 0, -100),
		VisibleEndTime:      time.Now().Local().AddDate(0, 0, 100),
		Rules:               rules,
		CreatedTime:         time.Now().Local().AddDate(0, 0, -1),
	}
	return &pct
}

func Test_ruleInit(t *testing.T) {
	Convey("validate rule init correctly", t, func() {
		So(len(issueJudges), ShouldEqual, 1)
		So(len(redeemJudges), ShouldEqual, 4)
		So(len(ruleBodyComposers), ShouldEqual, 4)
	})
}

func Test_validateTemplateRules(t *testing.T) {
	Convey("Given a coupon will no rule errors", t, func() {
		var pct = _aPublishedCouponType(_someRules())
		var atJudge *ApplyTimesJudge
		// atJudge = new(ApplyTimesJudge)
		atJudge = issueJudges["APPLY_TIMES"].(*ApplyTimesJudge)
		patchGuard := monkey.PatchInstanceMethod(reflect.TypeOf(atJudge), "JudgeTemplate", func(_ *ApplyTimesJudge, _ string, _ string, _ map[string]interface{}, _ *PublishedCouponType) error {
			return nil
		})

		Convey("Start validate", func() {
			errs, err := validateTemplateRules("", "", pct)
			Convey("Should no errors", func() {
				So(err, ShouldBeNil)
				So(len(errs), ShouldEqual, 0)
			})
		})

		patchGuard.Unpatch()
	})

	Convey("Given such env with no match rules", t, func() {
		var pct = _aPublishedCouponType(_someRules())
		patchGuard := monkey.Patch(findRule, func(ruleInternalID string) *Rule {
			return nil
		})

		Convey("Start validate", func() {
			_, err := validateTemplateRules("", "", pct)
			Convey("Should has ErrRuleNotFound", func() {
				So(err, ShouldEqual, &ErrRuleNotFound)
			})
		})

		patchGuard.Unpatch()
	})

	Convey("Given a template with no rules", t, func() {
		var pct = _aPublishedCouponType(nil)

		Convey("Start validate", func() {
			_, err := validateTemplateRules("", "", pct)
			Convey("Should has ErrRuleJudgeNotFound", func() {
				So(err, ShouldEqual, &ErrRuleJudgeNotFound)
			})
		})
	})

	Convey("Given an env that some judge will fail", t, func() {
		var pct = _aPublishedCouponType(_someRules())
		var atJudge *ApplyTimesJudge
		// atJudge = new(ApplyTimesJudge)
		atJudge = issueJudges["APPLY_TIMES"].(*ApplyTimesJudge)

		Convey("Assume ErrCouponRulesApplyTimeExpired", func() {

			patchGuard := monkey.PatchInstanceMethod(reflect.TypeOf(atJudge), "JudgeTemplate", func(_ *ApplyTimesJudge, _ string, _ string, _ map[string]interface{}, _ *PublishedCouponType) error {
				return &ErrCouponRulesApplyTimeExpired
			})
			errs, _ := validateTemplateRules("", "", pct)
			Convey("Should have ErrCouponRulesApplyTimeExpired", func() {
				So(len(errs), ShouldEqual, 1)
				So(errs[0], ShouldEqual, &ErrCouponRulesApplyTimeExpired)
			})
			patchGuard.Unpatch()
		})

		Convey("Assume ErrCouponRulesApplyTimesExceeded", func() {

			patchGuard := monkey.PatchInstanceMethod(reflect.TypeOf(atJudge), "JudgeTemplate", func(_ *ApplyTimesJudge, _ string, _ string, _ map[string]interface{}, _ *PublishedCouponType) error {
				return &ErrCouponRulesApplyTimesExceeded
			})
			errs, _ := validateTemplateRules("", "", pct)
			Convey("Should have ErrCouponRulesApplyTimesExceeded", func() {
				So(len(errs), ShouldEqual, 1)
				So(errs[0], ShouldEqual, &ErrCouponRulesApplyTimesExceeded)
			})
			patchGuard.Unpatch()
		})

	})
}

func Test_validateCouponRules(t *testing.T) {

	Convey("Given a coupon with no rules", t, func() {
		state := r.Intn(int(SUnknown))
		var p map[string]interface{}
		p = make(map[string]interface{}, 1)
		p[KeyBindingRuleProperties] = map[string]interface{}{}
		c := _aCoupon(base.RandString(4), "xxx", "yyy", defaultCouponTypeID, State(state), p)
		// pg1 := monkey.Patch(unmarshalCouponRules, func(_ string)([]*RuleRef, error) {
		// 	return make([]*RuleRef, 0), nil
		// })

		Convey("Start validate", func() {
			_, err := validateCouponRules(nil, "", c)
			Convey("Should has ErrRuleJudgeNotFound", func() {
				So(err, ShouldEqual, &ErrRuleJudgeNotFound)
			})
		})
		// pg1.Unpatch()
	})

	Convey("Given an env that some judge will fail", t, func() {
		state := r.Intn(int(SUnknown))
		var p map[string]interface{}
		p = make(map[string]interface{}, 1)
		var rrs = make(map[string]interface{}, 1)
		rid, rbd := _aREDEEM_TIMES_RuleRef(999)
		rrs[rid] = rbd
		p[KeyBindingRuleProperties] = rrs
		c := _aCoupon(base.RandString(4), "xxx", "yyy", defaultCouponTypeID, State(state), p)
		// c.Properties = p

		var rtJudge *RedeemTimesJudge
		rtJudge = redeemJudges["REDEEM_TIMES"].(*RedeemTimesJudge)

		Convey("Assume ErrCouponRulesRedeemTimesExceeded", func() {

			patchGuard := monkey.PatchInstanceMethod(reflect.TypeOf(rtJudge), "JudgeCoupon", func(_ *RedeemTimesJudge, _ *base.Requester, _ string, _ map[string]interface{}, _ *Coupon) error {
				// log.Println("=======monkey.JudgeCoupon=====")
				return &ErrCouponRulesRedeemTimesExceeded
			})
			// pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(c), "GetRules", func(_ *Coupon) map[string]interface{} {
			// log.Println("=======monkey.GetRules=====")
			// 	return rrs
			// })
			errs, _ := validateCouponRules(nil, "", c)
			// if nil != err {
			// 	log.Println("=======err is not nil =====")
			// 	log.Println(c.GetRules())
			// 	log.Println(err)
			// }
			// pg5.Unpatch()
			Convey("Should have ErrCouponRulesRedeemTimesExceeded", func() {
				So(len(errs), ShouldEqual, 1)
				So(errs[0], ShouldEqual, &ErrCouponRulesRedeemTimesExceeded)
			})
			patchGuard.Unpatch()
		})

	})

	Convey("Given an env that everything is Okay", t, func() {
		state := r.Intn(int(SUnknown))
		c := _aCoupon(base.RandString(4), "xxx", "yyy", defaultCouponTypeID, State(state), nil)
		// pg1 := monkey.Patch(unmarshalCouponRules, func(_ string)([]*RuleRef, error) {
		// 	rrs := make([]*RuleRef, 0, 1)
		// 	rrs = append(rrs, _aREDEEM_TIMES_RuleRef(1))
		// 	return rrs, nil
		// })

		Convey("Start validate", func() {
			var rtJudge *RedeemTimesJudge
			rtJudge = redeemJudges["REDEEM_TIMES"].(*RedeemTimesJudge)

			patchGuard := monkey.PatchInstanceMethod(reflect.TypeOf(rtJudge), "JudgeCoupon", func(_ *RedeemTimesJudge, _ *base.Requester, _ string, _ map[string]interface{}, _ *Coupon) error {
				return nil
			})
			errs, _ := validateCouponRules(nil, "", c)
			Convey("Should no error", func() {
				So(len(errs), ShouldEqual, 0)

			})
			patchGuard.Unpatch()
		})

		// pg1.Unpatch()
	})
}

func Test_marshalCouponRules(t *testing.T) {
	Convey("Given a rule will not be found", t, func() {
		patchGuard := monkey.Patch(findRule, func(_ string) *Rule {
			return nil
		})

		var rrs = make(map[string]map[string]interface{}, 4)
		rid, rbd := _aREDEEM_TIMES_RuleRef(1)
		rrs[rid] = rbd

		Convey("Start marshal", func() {
			_, err := marshalCouponRules(nil, "", rrs)
			Convey("Should has ErrRuleNotFound", func() {
				So(err, ShouldEqual, &ErrRuleNotFound)
			})
		})

		patchGuard.Unpatch()
	})

	Convey("Given a composer should return error", t, func() {
		var rrs = make(map[string]map[string]interface{}, 4)
		rid, rbd := _aREDEEM_TIMES_RuleRef(1)
		rrs[rid] = rbd

		var rdbComposer *RedeemTimesBodyComposer
		rdbComposer = ruleBodyComposers["REDEEM_TIMES"].(*RedeemTimesBodyComposer)

		patchGuard := monkey.PatchInstanceMethod(reflect.TypeOf(rdbComposer), "Compose", func(_ *RedeemTimesBodyComposer, _ *base.Requester, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
			// just pick an error for test, don't care if it is logical
			return nil, &ErrRequesterHasNoBrand
		})

		Convey("Start marshal", func() {
			_, err := marshalCouponRules(nil, "", rrs)
			Convey("Should has ErrRequesterHasNoBrand", func() {
				So(err, ShouldEqual, &ErrRequesterHasNoBrand)
			})
		})

		patchGuard.Unpatch()
	})

	Convey("Given an env that everything is Okay", t, func() {
		var rrs = make(map[string]map[string]interface{}, 4)
		rid, rbd := _aREDEEM_TIMES_RuleRef(1)
		rrs[rid] = rbd

		Convey("Start marshal", func() {
			m, err := marshalCouponRules(nil, "", rrs)
			Convey("Should has correct result", func() {
				So(err, ShouldBeNil)
				So(m["REDEEM_TIMES"], ShouldNotBeNil)
			})
		})
	})

}

// func Test_unmarshalCouponRules(t *testing.T) {
// 	Convey("Given a rule string with bad format (non json format)", t, func() {
// 		_, err := unmarshalCouponRules("bad_format")
// 		So(err, ShouldEqual, &ErrCouponRulesBadFormat)
// 	})

// 	Convey("Given a rule string with not RuleRef format", t, func() {
// 		_, err := unmarshalCouponRules(`{"hello" : "world"}`)
// 		So(err, ShouldEqual, &ErrCouponRulesBadFormat)
// 	})

// 	Convey("Given a rule string with RuleRef format", t, func() {
// 		rrfs, err := unmarshalCouponRules(` [ {"rule_id":"REDEEM_TIMES","rule_body":"abc"}]`)
// 		So(err, ShouldBeNil)
// 		So(len(rrfs), ShouldEqual,1)
// 		rrf := rrfs[0]
// 		So(rrf.RuleInternalID, ShouldEqual, "REDEEM_TIMES")
// 		So(rrf.RuleBody, ShouldEqual, "abc")
// 	})
// }

func Test_restRedeemTimes(t *testing.T) {
	state := r.Intn(int(SUnknown))
	var p map[string]interface{}
	p = make(map[string]interface{}, 1)
	var rrs = make(map[string]interface{}, 4)
	rid, rbd := _aREDEEM_TIMES_RuleRef(10000000)
	rrs[rid] = rbd
	p[KeyBindingRuleProperties] = rrs
	c := _aCoupon(base.RandString(4), "xxx", "yyy", defaultCouponTypeID, State(state), p)

	Convey("Given a env assume the db is down", t, func() {
		pg1 := monkey.Patch(getCouponTransactionCountWithType, func(_ string, _ TransType) (uint, error) {
			return 0, &ErrCouponRulesBadFormat
		})
		_, err := restRedeemTimes(c)
		So(err, ShouldEqual, &ErrCouponRulesBadFormat)
		pg1.Unpatch()
	})

	Convey("Given a env assume the coupon had been redeemed n times", t, func() {
		ts := r.Intn(10000)
		pg1 := monkey.Patch(getCouponTransactionCountWithType, func(_ string, _ TransType) (uint, error) {
			return uint(ts), nil
		})
		restts, err := restRedeemTimes(c)
		So(err, ShouldBeNil)
		So(restts, ShouldEqual, 10000000-ts)
		pg1.Unpatch()
	})

}
