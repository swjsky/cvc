package main

import (
//	"encoding/json"
	// "time"

	"loreal.com/dit/cmd/coupon-service/coupon"
	// "loreal.com/dit/cmd/coupon-service/rule"
)

// PreDefinedDataForDatabase 预定义的数据，
//TODO: 未来会重构掉
type preDefinedDataForDatabase struct {
	SupportedRules       []*coupon.Rule              `json:"supported_rules"`
	CouponTemplates      []*coupon.Template          `json:"coupon_templates"`
	PublishedCouponTypes []*coupon.PublishedCouponType `json:"published_coupon_types"`
}

