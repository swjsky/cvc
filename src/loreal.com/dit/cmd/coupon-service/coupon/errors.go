package coupon

import (
	"loreal.com/dit/cmd/coupon-service/base"
)

//ErrRuleJudgeNotFound - 找不到一个规则的校验。
var ErrRuleJudgeNotFound = base.ErrorWithCode{
	Code:    1000,
	Message: "rule judge not found",
}

//ErrRuleNotFound - 找不到一个规则的校验。
var ErrRuleNotFound = base.ErrorWithCode{
	Code:    1001,
	Message: "rule not found",
}

// ErrCouponRulesApplyTimesExceeded - 已经达到最大领用次数。
var ErrCouponRulesApplyTimesExceeded = base.ErrorWithCode{
	Code:    1002,
	Message: "coupon apply times exceeded",
}

// ErrCouponRulesApplyTimeExpired - 卡券已经过了申领期限。
var ErrCouponRulesApplyTimeExpired = base.ErrorWithCode{
	Code:    1003,
	Message: "the coupon applied has been expired",
}

// ErrCouponRulesBadFormat - 卡券附加的验证规则有错误。
var ErrCouponRulesBadFormat = base.ErrorWithCode{
	Code:    1004,
	Message: "the coupon has a bad formated rules",
}

// ErrCouponRulesRedemptionNotStart - 卡券还没开始核销。
var ErrCouponRulesRedemptionNotStart = base.ErrorWithCode{
	Code:    1005,
	Message: "the coupon has not start the redemption",
}

// ErrCouponRulesRedeemTimesExceeded - 已经达到最大领用次数。
var ErrCouponRulesRedeemTimesExceeded = base.ErrorWithCode{
	Code:    1006,
	Message: "coupon redeem times exceeded",
}

// ErrCouponRulesNoRedeemTimes - 已经达到最大领用次数。
var ErrCouponRulesNoRedeemTimes = base.ErrorWithCode{
	Code:    1007,
	Message: "coupon has no redeem times rule",
}

// ErrCouponRulesRedemptionExpired - 卡券核销已经过期。
var ErrCouponRulesRedemptionExpired = base.ErrorWithCode{
	Code:    1008,
	Message: "the coupon is expired",
}

// ErrCouponRulesUnsuportTimeUnit - 不支持的时间单位。
var ErrCouponRulesUnsuportTimeUnit = base.ErrorWithCode{
	Code:    1009,
	Message: "the coupon redeem time unit unsupport",
}

//ErrCouponTemplateNotFound - 签发新的Coupon时，找不到Coupon的模板的类型
var ErrCouponTemplateNotFound = base.ErrorWithCode{
	Code:    1100,
	Message: "coupon template not found",
}

//ErrCouponIDInvalid - 签发新的Coupon时，用户id或者卡券类型不合法
var ErrCouponIDInvalid = base.ErrorWithCode{
	Code:    1200,
	Message: "coupon id is invalid",
}

//ErrCouponNotFound - 没找到Coupon时
var ErrCouponNotFound = base.ErrorWithCode{
	Code:    1201,
	Message: "coupon not found",
}

//ErrCouponIsNotActive - 没找到Coupon时
var ErrCouponIsNotActive = base.ErrorWithCode{
	Code:    1202,
	Message: "coupon is not active",
}

//ErrCouponWasRedeemed - 没找到Coupon时
var ErrCouponWasRedeemed = base.ErrorWithCode{
	Code:    1203,
	Message: "coupon was redeemed",
}

//ErrCouponTooMuchToRedeem - 没找到Coupon时
var ErrCouponTooMuchToRedeem = base.ErrorWithCode{
	Code:    1204,
	Message: "too much coupons to redeem",
}

//ErrCouponWrongConsumer - 没找到Coupon时
var ErrCouponWrongConsumer = base.ErrorWithCode{
	Code:    1205,
	Message: "the coupon's owner is not the provided consumer",
}

//ErrConsumerIDAndCouponTypeIDInvalid - 签发新的Coupon时，用户id或者卡券类型不合法
var ErrConsumerIDAndCouponTypeIDInvalid = base.ErrorWithCode{
	Code:    1300,
	Message: "consumer id or coupon type is invalid",
}

//ErrConsumerIDInvalid - 签发新的Coupon时，用户id或者卡券类型不合法
var ErrConsumerIDInvalid = base.ErrorWithCode{
	Code:    1301,
	Message: "consumer id is invalid",
}

//ErrConsumerIDsAndRefIDsMismatch - 消费者的RefID和ID数量不匹配
var ErrConsumerIDsAndRefIDsMismatch = base.ErrorWithCode{
	Code:    1302,
	Message: "consumer ids and the ref ids mismatch",
}

//ErrRequesterForbidden - 没找到Coupon时
var ErrRequesterForbidden = base.ErrorWithCode{
	Code:    1400,
	Message: "requester was forbidden to do this action",
}

//ErrRequesterHasNoBrand - 没找到Coupon时
var ErrRequesterHasNoBrand = base.ErrorWithCode{
	Code:    1401,
	Message: "requester has no brand information",
}

//ErrRedeemWithDiffBrand - 没找到Coupon时
var ErrRedeemWithDiffBrand = base.ErrorWithCode{
	Code:    1402,
	Message: "redeem coupon with different brand",
}
