package base

import (
	"github.com/chenhg5/collection"
	"github.com/dgrijalva/jwt-go"
)

const ROLE_COUPON_ISSUER string = "coupon_issuer"
const ROLE_COUPON_REDEEMER string = "coupon_redeemer"
const ROLE_COUPON_LISTENER string = "coupon_listener"

// Requester 解析数据
type Requester struct { // token里面添加用户信息，验证token后可能会用到用户信息
	jwt.StandardClaims
	UserID string                `json:"preferred_username"` //TODO: 能否改成username？
	Roles  map[string]([]string) `json:"realm_access"`
	Brand  string                `json:"brand"`
}

// HasRole 检查请求者是否有某个角色
func (r *Requester) HasRole(role string) bool {
	// if	!collection.Collect(requester.Roles).Has("roles") {
	// 		return false
	// }
	roles := r.Roles["roles"]
	if nil == roles {
		return false
	}
	if !collection.Collect(roles).Contains(role) {
		return false
	}
	return true
}
