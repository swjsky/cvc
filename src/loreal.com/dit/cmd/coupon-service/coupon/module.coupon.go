package coupon

import (
	"encoding/json"
	"log"
	"time"

	"loreal.com/dit/utils"
	// "loreal.com/dit/cmd/coupon-service/rule"
)

// KeyBindingRuleProperties 生成的Coupon包含类型为map的Properties字段，用来保存多样化数据。KeyBindingRuleProperties对应的值是该券在兑换时需要满足的条件。
const KeyBindingRuleProperties string = "binding_rule_properties"

// Template 卡券的模板
type Template struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Creator     string                 `json:"creator"`
	Rules       map[string]interface{} `json:"rules"`
	CreatedTime time.Time              `json:"created_time,omitempty" type:"DATETIME" default:"datetime('now','localtime')"`
	UpdatedTime time.Time              `json:"updated_time,omitempty" type:"DATETIME"`
	DeletedTime time.Time              `json:"deleted_time,omitempty" type:"DATETIME"`
}

// CTState 卡券类型的状态定义
type CTState int32

// 卡券的状态
const (
	CTSActive State = iota
	CTSRevoked
	CTSUnknown
)

// PublishedCouponType 已经发布的卡券类型
type PublishedCouponType struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	TemplateID          string            `json:"template_id"`
	Description         string            `json:"description"`
	InternalDescription string            `json:"internal_description"`
	State               CTState           `json:"state"`
	Publisher           string            `json:"publisher"`
	VisibleStartTime    time.Time         `json:"visible_start_time" type:"DATETIME"`
	VisibleEndTime      time.Time         `json:"visible_end_time"`
	StrRules            map[string]string `json:"rules"`
	Rules               map[string]map[string]interface{}
	CreatedTime         time.Time `json:"created_time" type:"DATETIME" default:"datetime('now','localtime')"`
	DeletedTime         time.Time `json:"deleted_time" type:"DATETIME"`
}

// InitRules //TODO 未来会重构掉
func (t *PublishedCouponType) InitRules() {
	t.Rules = map[string]map[string]interface{}{}
	for k, v := range t.StrRules {
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(v), &obj)
		if nil == err {
			t.Rules[k] = obj
		} else {
			log.Panic(err)

		}
	}
}

// State 卡券的状态类型定义
type State int32

// 卡券的状态
// TODO [HUBIN]: 增加 SExpired 状态
const (
	SActive State = iota //如果非一次兑换类型，那么在有效兑换次数内，仍然是active
	SRevoked
	SDeleteCoupon
	SRedeemed //无论多次还是一次，全部用完后置为该状态
	SExpired
	SUnknown
)

// Coupon 用来封装一个Coupon实体的结构
type Coupon struct {
	ID            string
	CouponTypeID  string
	ConsumerID    string
	ConsumerRefID string
	ChannelID     string
	State         State
	Properties    map[string]interface{}
	CreatedTime   *time.Time
	Transactions  []*Transaction
}

// CreatedTimeToLocal 使用localtime
func (c *Coupon) CreatedTimeToLocal() {
	l := c.CreatedTime.Local()
	c.CreatedTime = &l
}

// RedeemedCoupons 传递被核销卡券信息的结构，
type RedeemedCoupons struct {
	ExtraInfo string    `json:"extrainfo,omitempty"`
	Coupons   []*Coupon `json:"coupons,omitempty"`
}

// TransType 卡券被操作的状态类型
type TransType int32

// 卡券被操作的种类
const (
	TTIssueCoupon TransType = iota
	TTDeactiveCoupon
	TTDeleteCoupon
	TTRedeemCoupon //可多次存在
	TTExpired
	TTUnknownTransaction
)

// Transaction 用来封装一次Coupon的状态变动
type Transaction struct {
	ID          string
	CouponID    string
	ActorID     string
	TransType   TransType
	ExtraInfo   string
	CreatedTime time.Time
}

// EncryptExtraInfo 给ExtraInfo 使用AES256加密
func (t *Transaction) EncryptExtraInfo() string {
	return utils.AES256URLEncrypt(t.ExtraInfo, encryptKey)
}

// DecryptExtraInfo 给ExtraInfo 使用AES256解密
func (t *Transaction) DecryptExtraInfo(emsg string) error {
	p, e := utils.AES256URLDecrypt(emsg, encryptKey)
	if nil != e {
		return e
	}
	t.ExtraInfo = p
	return nil
}

// CreatedTimeToLocal 给ExtraInfo 使用AES256解密
func (t *Transaction) CreatedTimeToLocal() {
	l := t.CreatedTime.Local()
	t.CreatedTime = l
}
