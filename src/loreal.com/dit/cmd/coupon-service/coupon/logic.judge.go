package coupon

import (
	// "encoding/json"
	"log"
	"time"

	"loreal.com/dit/cmd/coupon-service/base"

	"github.com/jinzhu/now"
	"github.com/mitchellh/mapstructure"
)

// var daysMap = map[string]int {
// 	"YEAR" : 366, 		//理想的是根据是否闰年来计算

// }

// TemplateJudge 发卡券时用来验证是否符合rules
type TemplateJudge interface {
	// JudgeTemplate 验证模板
	JudgeTemplate(consumerID string, couponTypeID string, ruleBody map[string]interface{}, pct *PublishedCouponType) error
}

// Judge 兑换卡券时用来验证是否符合rules
type Judge interface {
	// JudgeCoupon 验证模板
	JudgeCoupon(requester *base.Requester, consumerID string, ruleBody map[string]interface{}, c *Coupon) error
}

// RedeemPeriodJudge 验证有效期
type RedeemPeriodJudge struct {
}

// RedeemInCurrentNatureTimeUnitJudge 验证自然月，季度，年度有效期
type RedeemInCurrentNatureTimeUnitJudge struct {
}

//ApplyTimesJudge 验证领用次数
type ApplyTimesJudge struct {
}

//RedeemTimesJudge 验证兑换次数
type RedeemTimesJudge struct {
}

//RedeemBySameBrandJudge 验证是否同品牌兑换
type RedeemBySameBrandJudge struct {
}

// TODO: 重构rule结构实现这些接口，统一处理。

// JudgeCoupon 验证有效期
// TODO: 未来加上DAY, WEEK, SEARON, YEAR 等
func (*RedeemInCurrentNatureTimeUnitJudge) JudgeCoupon(requester *base.Requester, consumerID string, ruleBody map[string]interface{}, c *Coupon) error {

	var ntu natureTimeUnit
	// var err error
	var start time.Time
	var end time.Time

	if err := mapstructure.Decode(ruleBody, &ntu); err != nil {
		return &ErrCouponRulesBadFormat
	}

	// TODO: 支持季度，年啥的。
	if base.IsBlankString(ntu.Unit) || ntu.Unit != "MONTH" {
		return &ErrCouponRulesUnsuportTimeUnit
	}

	switch ntu.Unit {
	case "MONTH":
		{
			ctMonth := now.With(*c.CreatedTime)
			start = ctMonth.BeginningOfMonth()
			end = ctMonth.EndOfMonth().AddDate(0, 0, ntu.EndInAdvance*-1)
		}
	}
	n := time.Now()
	if n.Before(start) {
		return &ErrCouponRulesRedemptionNotStart
	}

	if n.After(end) {
		return &ErrCouponRulesRedemptionExpired
	}

	return nil
}

// JudgeCoupon 验证有效期
func (*RedeemPeriodJudge) JudgeCoupon(requester *base.Requester, consumerID string, ruleBody map[string]interface{}, c *Coupon) error {

	var ts timeSpan
	var err error
	var start time.Time
	var end time.Time
	// err := json.Unmarshal([]byte(jsonString), &ts)
	// if err != nil {
	// 	log.Println(err)
	// 	return &ErrCouponRulesBadFormat
	// }

	// var startTime, endTime string
	// var ok bool
	// if	startTime, ok = ruleBody["startTime"].(string); !ok {
	// 	return &ErrCouponRulesBadFormat
	// }
	if err := mapstructure.Decode(ruleBody, &ts); err != nil {
		return &ErrCouponRulesBadFormat
	}

	if !base.IsBlankString(ts.StartTime) {
		start, err = time.Parse(timeLayout, ts.StartTime)
		if nil != err {
			log.Println(err)
			return &ErrCouponRulesBadFormat
		}
	}

	// if	endTime, ok = ruleBody["endTime"].(string); !ok {
	// 	return &ErrCouponRulesBadFormat
	// }

	if !base.IsBlankString(ts.EndTime) {
		end, err = time.Parse(timeLayout, ts.EndTime)
		if nil != err {
			log.Println(err)
			return &ErrCouponRulesBadFormat
		}
	}

	if time.Now().Before(start) {
		return &ErrCouponRulesRedemptionNotStart
	}

	if time.Now().After(end) {
		return &ErrCouponRulesRedemptionExpired
	}

	return nil
}

// JudgeCoupon 验证兑换次数
func (*RedeemTimesJudge) JudgeCoupon(requester *base.Requester, consumerID string, ruleBody map[string]interface{}, c *Coupon) error {

	var rt redeemTimes
	// err := json.Unmarshal([]byte(jsonString), &rt)
	// if err != nil {
	// 	log.Println(err)
	// return err
	// }

	tas, err := getCouponTransactionsWithType(c.ID, TTRedeemCoupon)
	if err != nil {
		return err
	}

	if nil == tas {
		return nil
	}

	// var i int
	// var ok bool
	// if	i, ok = ruleBody["times"].(int); !ok {
	// 		return &ErrCouponRulesBadFormat
	// }

	if err := mapstructure.Decode(ruleBody, &rt); err != nil {
		return &ErrCouponRulesBadFormat
	}

	if len(tas) >= int(rt.Times) {
		return &ErrCouponRulesRedeemTimesExceeded
	}
	return nil
}

// JudgeCoupon 验证只能同品牌兑换
func (*RedeemBySameBrandJudge) JudgeCoupon(requester *base.Requester, consumerID string, ruleBody map[string]interface{}, c *Coupon) error {
	if base.IsEmptyString(requester.Brand) {
		return &ErrRequesterHasNoBrand
	}

	var brand sameBrand
	// err := json.Unmarshal([]byte(jsonString), &brand)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

	// var brand string
	// var ok bool
	if err := mapstructure.Decode(ruleBody, &brand); err != nil {
		return &ErrCouponRulesBadFormat
	}

	// if	brand, ok = ruleBody["brand"].(string); !ok {
	// 	return &ErrCouponRulesBadFormat
	// }

	if requester.Brand != brand.Brand {
		return &ErrRedeemWithDiffBrand
	}

	return nil
}

// JudgeTemplate 验证领用次数
func (*ApplyTimesJudge) JudgeTemplate(consumerID string, couponTypeID string, ruleBody map[string]interface{}, pct *PublishedCouponType) error {
	coupons, err := getCoupons(consumerID, couponTypeID)
	if err != nil {
		return err
	}

	var at applyTimes
	// err2 := json.Unmarshal([]byte(jsonString), &at)
	// if err2 != nil {
	// 	log.Println(err2)
	// 	return err2
	// }

	// var i int
	// var ok bool

	if err := mapstructure.Decode(ruleBody, &at); err != nil {
		return &ErrCouponRulesBadFormat
	}

	// if	i, ok = ruleBody["inDays"].(int); !ok {
	// 	return &ErrCouponRulesBadFormat
	// }

	if at.InDays > 0 {
		var expiredTime = pct.CreatedTime.AddDate(0, 0, int(at.InDays))
		if !time.Now().Before(expiredTime) {
			return &ErrCouponRulesApplyTimeExpired
		}
	}

	// if	i, ok = ruleBody["times"].(int); !ok {
	// 	return &ErrCouponRulesBadFormat
	// }

	if at.Times <= uint(len(coupons)) {
		return &ErrCouponRulesApplyTimesExceeded
	}

	return nil
}

// JudgeNTUExpired - To validate experiation status of coupon for NTU type.
func JudgeNTUExpired(requester *base.Requester, consumerID string, ruleBody map[string]interface{}, c *Coupon) (bool, error) {
	var ntu natureTimeUnit
	var end time.Time

	if err := mapstructure.Decode(ruleBody, &ntu); err != nil {
		return false, &ErrCouponRulesBadFormat
	}

	if base.IsBlankString(ntu.Unit) || ntu.Unit != "MONTH" {
		return false, &ErrCouponRulesUnsuportTimeUnit
	}

	switch ntu.Unit {
	case "MONTH":
		{
			ctMonth := now.With(*c.CreatedTime)
			end = ctMonth.EndOfMonth().AddDate(0, 0, ntu.EndInAdvance*-1)
		}
	}
	n := time.Now()

	if n.After(end) {
		return true, nil
	}

	return false, nil
}

// JudgePeriodExpired - To validate experiation status of coupon for normal period type.
func JudgePeriodExpired(requester *base.Requester, consumerID string, ruleBody map[string]interface{}, c *Coupon) (bool, error) {
	var ts timeSpan
	var err error
	var end time.Time

	if err := mapstructure.Decode(ruleBody, &ts); err != nil {
		return false, &ErrCouponRulesBadFormat
	}

	if !base.IsBlankString(ts.EndTime) {
		end, err = time.Parse(timeLayout, ts.EndTime)
		if nil != err {
			log.Println(err)
			return false, &ErrCouponRulesBadFormat
		}
	}

	if time.Now().After(end) {
		return true, nil
	}

	return false, nil
}
