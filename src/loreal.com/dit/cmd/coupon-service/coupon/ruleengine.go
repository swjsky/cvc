package coupon

import (
	// "encoding/json"
	// "fmt"
	"log"

	"github.com/mitchellh/mapstructure"
	"loreal.com/dit/cmd/coupon-service/base"
)

var supportedRules []*Rule

var issueJudges map[string]TemplateJudge
var ruleBodyComposers map[string]BodyComposer
var redeemJudges map[string]Judge

// Init 初始化规则的一些基础数据
// TODO: 这些可以通过配置文件进行，避免未来修改程序后才能部署
func ruleInit(rules []*Rule) {
	supportedRules = rules
	issueJudges = make(map[string]TemplateJudge)
	issueJudges["APPLY_TIMES"] = new(ApplyTimesJudge)

	redeemJudges = make(map[string]Judge)
	redeemJudges["REDEEM_PERIOD_WITH_OFFSET"] = new(RedeemPeriodJudge)
	redeemJudges["REDEEM_TIMES"] = new(RedeemTimesJudge)
	redeemJudges["REDEEM_BY_SAME_BRAND"] = new(RedeemBySameBrandJudge)
	redeemJudges["REDEEM_IN_CURRENT_NATURE_MONTH_SEASON_YEAR"] = new(RedeemInCurrentNatureTimeUnitJudge)

	ruleBodyComposers = make(map[string]BodyComposer)
	ruleBodyComposers["REDEEM_PERIOD_WITH_OFFSET"] = new(RedeemPeriodWithOffsetBodyComposer)
	ruleBodyComposers["REDEEM_IN_CURRENT_NATURE_MONTH_SEASON_YEAR"] = new(RedeemInCurrentNatureTimeUnitBodyComposer)
	ruleBodyComposers["REDEEM_TIMES"] = new(RedeemTimesBodyComposer)
	ruleBodyComposers["REDEEM_BY_SAME_BRAND"] = new(RedeemBySameBrandBodyComposer)
}

// ValidateTemplateRules 发券时验证是否可以领用
// 目前要求至少配置一个领用规则
func validateTemplateRules(consumerID string, couponTypeID string, pct *PublishedCouponType) ([]error, error) {
	var errs []error = make([]error, 0)

	var judgeTimes int = 0
	for ruleInternalID, tempRuleBody := range pct.Rules {
		var rule = findRule(ruleInternalID)
		if nil == rule {
			return nil, &ErrRuleNotFound
		}
		if judge, ok := issueJudges[ruleInternalID]; ok {
			judgeTimes++
			var err = judge.JudgeTemplate(consumerID, couponTypeID, tempRuleBody, pct)
			if nil != err {
				errs = append(errs, err)
			}
			continue
		}

	}

	if 0 == judgeTimes {
		return nil, &ErrRuleJudgeNotFound
	}

	return errs, nil
}

// ValidateCouponRules 在redeem时需要验证卡券的规则
// 目前要求至少配置一个兑换规则
func validateCouponRules(requester *base.Requester, consumerID string, c *Coupon) ([]error, error) {
	var errs []error = make([]error, 0)

	var judgeTimes int = 0
	// var rulesString = c.Properties[KeyBindingRuleProperties]
	// ruleBodyRefs, err := unmarshalCouponRules(rulesString)
	// if nil != err {
	// 	return nil, err
	// }
	// log.Println(c.GetRules())
	// for ruleInternalID, ruleBody := range c.Properties[KeyBindingRuleProperties].(map[string]interface {}) {
	for ruleInternalID, ruleBody := range c.GetRules() {
		// TODO: 未来性能优化，考虑这个findRule去掉
		var rule = findRule(ruleInternalID)
		if nil == rule {
			return nil, &ErrRuleNotFound
		}
		if judge, ok := redeemJudges[ruleInternalID]; ok {
			judgeTimes++
			var err = judge.JudgeCoupon(requester, consumerID, ruleBody.(map[string]interface{}), c)
			if nil != err {
				errs = append(errs, err)
			}
			continue
		}
	}

	// for _, ruleRef := range ruleBodyRefs {
	// 	var ruleInternalID = ruleRef.RuleInternalID
	// 	var ruleBody = ruleRef.RuleBody
	// 	var rule = findRule(ruleInternalID)
	// 	if nil == rule {
	// 		return nil, &ErrRuleNotFound
	// 	}
	// 	if judge, ok := redeemJudges[rule.InternalID]; ok {
	// 		judgeTimes++
	// 		var err = judge.JudgeCoupon(requester, consumerID, ruleBody, c)
	// 		if nil != err {
	// 			errs = append(errs, err)
	// 		}
	// 		continue
	// 	}
	// }
	if 0 == judgeTimes {
		return nil, &ErrRuleJudgeNotFound
	}

	return errs, nil
}

// validateCouponExpired
func validateCouponExpired(requester *base.Requester, consumerID string, c *Coupon) (bool, error) {
	for ruleInternalID, ruleBody := range c.GetRules() {
		var rule = findRule(ruleInternalID)
		if nil == rule {
			return false, &ErrRuleNotFound
		}
		if _, ok := redeemJudges[ruleInternalID]; ok {
			if ruleInternalID == "REDEEM_IN_CURRENT_NATURE_MONTH_SEASON_YEAR" {
				expired, err := JudgeNTUExpired(requester, consumerID, ruleBody.(map[string]interface{}), c)
				return expired, err
			}

			if ruleInternalID == "REDEEM_PERIOD_WITH_OFFSET" {
				expired, err := JudgePeriodExpired(requester, consumerID, ruleBody.(map[string]interface{}), c)
				return expired, err
			}
		}
	}
	return false, &ErrRuleNotFound
}

func findRule(ruleInternalID string) *Rule {
	for _, rule := range supportedRules {
		if rule.InternalID == ruleInternalID {
			return rule
		}
	}
	return nil
}

// marshalCouponRules 发券时生成最终的规则字符串，用来在核销时验证
func marshalCouponRules(requester *base.Requester, couponTypeID string, rules map[string]map[string]interface{}) (map[string]interface{}, error) {
	var ruleBodyMap map[string]interface{} = make(map[string]interface{}, 0)
	for ruleInternalID, tempRuleBody := range rules {
		var rule = findRule(ruleInternalID)
		if nil == rule {
			return nil, &ErrRuleNotFound
		}

		if composer, ok := ruleBodyComposers[ruleInternalID]; ok {
			ruleBody, err := composer.Compose(requester, couponTypeID, tempRuleBody)
			if nil != err {
				log.Println(err)
				return nil, err
			}
			ruleBodyMap[ruleInternalID] = ruleBody
		}
	}
	return ruleBodyMap, nil
	// jsonBytes, err := json.Marshal(ruleBodyRefs)
	// if nil != err {
	// 	log.Println(err)
	// 	return "", err
	// }

	// return string(jsonBytes), nil
}

// unmarshalCouponRules 兑换时将卡券附带的规则解析成[]*RuleRef
// func unmarshalCouponRules(rulesString string) ([]*RuleRef, error) {
// 	var ruleBodyRefs []*RuleRef = make([]*RuleRef, 0)
// 	err := json.Unmarshal([]byte(rulesString), &ruleBodyRefs)
// 	if nil != err {
// 		log.Println(err)
// 		return nil, &ErrCouponRulesBadFormat
// 	}
// 	return ruleBodyRefs, nil
// }

// restRedeemTimes 查询剩下redeem的次数
func restRedeemTimes(c *Coupon) (uint, error) {
	// for ruleInternalID, ruleBody := range c.Properties[KeyBindingRuleProperties].(map[string]interface {}) {
	for ruleInternalID, ruleBody := range c.GetRules() {
		if "REDEEM_TIMES" == ruleInternalID {
			// var count uint
			count, err := getCouponTransactionCountWithType(c.ID, TTRedeemCoupon)
			if nil != err {
				return 0, err
			}
			var rt redeemTimes
			if err = mapstructure.Decode(ruleBody, &rt); err != nil {
				return 0, &ErrCouponRulesBadFormat
			}

			// comparing first, to avoid negative result causes uint type out of bound
			if uint(rt.Times) >= count {
				return uint(rt.Times) - count, nil
			}

			return 0, nil
		}
	}

	// TODO: 未来考虑此处查找优化，提升效率
	// for _, ruleRef := range ruleBodyRefs {
	// 	if ruleRef.RuleInternalID == "REDEEM_TIMES" {
	// 		var ruleBody = ruleRef.RuleBody
	// 		var rt redeemTimes
	// 		err := json.Unmarshal([]byte(ruleBody), &rt)
	// 		if err != nil {
	// 			log.Println(err)
	// 			return 0, err
	// 		}

	// 		// 查询已经兑换次数
	// 		var count uint
	// 		count, err = getCouponTransactionCountWithType(c.ID, TTRedeemCoupon)
	// 		if nil != err {
	// 			return 0, err
	// 		}

	// 		return rt.Times - count, nil
	// 	}
	// }

	// 卡券没有设置兑换次数限制
	return 0, &ErrCouponRulesNoRedeemTimes
}
