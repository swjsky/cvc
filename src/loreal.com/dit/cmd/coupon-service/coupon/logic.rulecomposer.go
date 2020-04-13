package coupon

import (
	// "encoding/json"
	// "log"
	"time"

	// "loreal.com/dit/cmd/coupon-service/coupon"
	"loreal.com/dit/cmd/coupon-service/base"

	"github.com/mitchellh/mapstructure"
)

const timeLayout string = "2006-01-02 15:04:05 -07"

// BodyComposer 发卡时生成rule的body，用来存在coupon中
type BodyComposer interface {
	Compose(requester *base.Requester, couponTypeID string, ruleBody map[string]interface{}) (map[string]interface{}, error)
}

// RedeemTimesBodyComposer 生成兑换次数的内容
type RedeemTimesBodyComposer struct {
}

// RedeemBySameBrandBodyComposer 生成兑换次数的内容
type RedeemBySameBrandBodyComposer struct {
}

// RedeemPeriodWithOffsetBodyComposer 生成有效期的内容
type RedeemPeriodWithOffsetBodyComposer struct {
}

// RedeemInCurrentNatureTimeUnitBodyComposer 生成有效期的内容
type RedeemInCurrentNatureTimeUnitBodyComposer struct {
}

// Compose 生成兑换次数的内容
// 规则体像 {"times": 1024}
func (*RedeemTimesBodyComposer) Compose(requester *base.Requester, couponTypeID string, ruleBody map[string]interface{}) (map[string]interface{}, error) {
	return ruleBody, nil
}

// Compose 生成由同品牌才能兑换的内容
// 规则体像 {"brand":"Lancome"}
func (*RedeemBySameBrandBodyComposer) Compose(requester *base.Requester, couponTypeID string, ruleBody map[string]interface{}) (map[string]interface{}, error) {
	if base.IsEmptyString(requester.Brand) {
		return nil, &ErrRequesterHasNoBrand
	}
	var brand = make(map[string]interface{})
	brand["brand"] = requester.Brand
	// var brand = sameBrand{
	// 	Brand: requester.Brand,
	// }
	return brand, nil
	// return (&brand).(map[string]interface{}), nil
}

// Compose 生成卡券有效期的规则体
// 模板内的规则类似这样的格式： {"offSetFromAppliedDay": 14,"timeSpan": 365}， 表示从领用日期延期14天后生效可以兑换，截止日期是14+365天内。
// 如果offSetFromAppliedDay=0，则当天生效。如果 timeSpan=0，则无过期时间。
// 时间单位都是天。
// 生成卡券后的规则：{ "startTime": "2020-02-14T00:00:00+08:00",  "endTime": null }， 表示从2020-02-14日 0点开始，无过期时间。
func (*RedeemPeriodWithOffsetBodyComposer) Compose(requester *base.Requester, couponTypeID string, ruleBody map[string]interface{}) (map[string]interface{}, error) {
	var offset offsetSpan

	if err := mapstructure.Decode(ruleBody, &offset); err != nil {
		return nil, &ErrCouponRulesBadFormat
	}
	var span = make(map[string]interface{})

	// var i int
	// var ok bool
	// if	i, ok = ruleBody["offSetFromAppliedDay"].(int); !ok {
	// 	return nil, &ErrCouponRulesBadFormat
	// }

	var st = time.Now().Local().AddDate(0, 0, int(offset.OffSetFromAppliedDay))
	// 时分秒清零
	// st = Time.Date(st.Year())
	span["startTime"] = st.Format(timeLayout)

	// if	i, ok = ruleBody["timeSpan"].(int); !ok {
	// 	return nil, &ErrCouponRulesBadFormat
	// }

	ts := offset.TimeSpan
	if 0 != ts {
		var et = st.AddDate(0, 0, int(ts))
		span["endTime"] = et.Format(timeLayout)
	}

	return span, nil
}

// Compose 生成卡券有效期的规则体，基于自然月，季度，年等。
// 模板内的规则类似这样的格式： {"unit": "MONTH", "endInAdvance": 5}， 表示领用当月生效，但在当月结束前5天过期。
// endInAdvance 的单位是 天， 默认为 0。
func (*RedeemInCurrentNatureTimeUnitBodyComposer) Compose(requester *base.Requester, couponTypeID string, ruleBody map[string]interface{}) (map[string]interface{}, error) {
	return ruleBody, nil
}
