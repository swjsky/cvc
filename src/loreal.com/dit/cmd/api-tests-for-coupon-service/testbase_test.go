package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	// "fmt"
	"math/rand"

	// "reflect"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
)

var r *rand.Rand = rand.New(rand.NewSource(time.Now().Unix()))

const defaultCouponTypeID string = "678719f5-44a8-4ac8-afd0-288d2f14daf8"
const anotherCouponTypeID string = "dff0710e-f5af-4ecf-a4b5-cc5599d98030"

var dbConnection *sql.DB

func TestMain(m *testing.M) {
	_setUp()
	m.Run()
	_tearDown()
}

func _setUp() {
	client := &http.Client{}
	// 激活测试数据，未来会重构
	req, err := http.NewRequest("GET", baseurl+"/api/apitester", strings.NewReader("name=cjb"))
	if err != nil {
		fmt.Println(err.Error())
	}
	req.Header.Set("Authorization", jwtyhl5000d)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(resp.Status)

	dbConnection, err = sql.Open("sqlite3", "../coupon-service/data/data.db?cache=shared&mode=rwc")
	if err != nil {
		panic(err)
	}
}

func _tearDown() {
}

// func _aCoupon(consumerID string, consumerRefID string, channelID string, couponTypeID string, state State, p map[string]interface{}) *Coupon {
// 	lt := time.Now().Local()

// 	var c = Coupon{
// 		ID:            uuid.New().String(),
// 		CouponTypeID:  couponTypeID,
// 		ConsumerID:    consumerID,
// 		ConsumerRefID: consumerRefID,
// 		ChannelID:     channelID,
// 		State:         state,
// 		Properties:    p,
// 		CreatedTime:   &lt,
// 	}
// 	return &c
// }

// func _someCoupons(consumerID string, consumerRefID string, channelID string, couponTypeID string) []*Coupon {
// 	count := r.Intn(10) + 1
// 	cs := make([]*Coupon, 0, count)
// 	for i := 0; i < count; i++ {
// 		state := r.Intn(int(SUnknown))
// 		var p map[string]interface{}
// 		p = make(map[string]interface{}, 1)
// 		p["the_key"] = "the value"
// 		cs = append(cs, _aCoupon(consumerID, consumerRefID, channelID, couponTypeID, State(state), p))
// 	}
// 	return cs
// }

// func _aTransaction(actorID string, couponID string, tt TransType, extraInfo string) *Transaction {
// 	var t = Transaction{
// 		ID:          uuid.New().String(),
// 		CouponID:    couponID,
// 		ActorID:     actorID,
// 		TransType:   tt,
// 		ExtraInfo:   extraInfo,
// 		CreatedTime: time.Now().Local(),
// 	}
// 	return &t
// }

// func _someTransaction(actorID string, couponID string, extraInfo string) []*Transaction {
// 	count := r.Intn(10) + 1
// 	ts := make([]*Transaction, 0, count)
// 	for i := 0; i < count; i++ {
// 		tt := r.Intn(int(TTUnknownTransaction))
// 		ts = append(ts, _aTransaction(actorID, couponID, TransType(tt), extraInfo))
// 	}
// 	return ts
// }

// func _aRequester(userID string, roles []string, brand string) *base.Requester {
// 	var requester base.Requester
// 	requester.UserID = userID
// 	requester.Roles = map[string]([]string){
// 		"roles": roles,
// 	}
// 	requester.Brand = brand
// 	return &requester
// }

func _validateCouponWithTypeA(c *httpexpect.Object, cid string, cref string, channelID string, state int) {
	__validateBasicCoupon(c, cid, cref, channelID, state)
	c.ContainsKey("CouponTypeID").ValueEqual("CouponTypeID", typeA)
	ps := c.ContainsKey("Properties").Value("Properties").Object()
	bind := ps.ContainsKey("binding_rule_properties").Value("binding_rule_properties").Object()
	nature := bind.ContainsKey("REDEEM_IN_CURRENT_NATURE_MONTH_SEASON_YEAR").Value("REDEEM_IN_CURRENT_NATURE_MONTH_SEASON_YEAR").Object()
	nature.ContainsKey("unit").ValueEqual("unit", "MONTH")
	nature.ContainsKey("endInAdvance").ValueEqual("endInAdvance", 0)
	rdtimes := bind.ContainsKey("REDEEM_TIMES").Value("REDEEM_TIMES").Object()
	rdtimes.ContainsKey("times").ValueEqual("times", 1)
}

func _validateCouponWithTypeB(c *httpexpect.Object, cid string, cref string, channelID string, state int) {
	__validateBasicCoupon(c, cid, cref, channelID, state)
	c.ContainsKey("CouponTypeID").ValueEqual("CouponTypeID", typeB)

	// ps := c.ContainsKey("Properties").Value("Properties").Object()
	// bind := ps.ContainsKey("binding_rule_properties").Value("binding_rule_properties").Object()
	// nature := bind.ContainsKey("REDEEM_IN_CURRENT_NATURE_MONTH_SEASON_YEAR").Value("REDEEM_IN_CURRENT_NATURE_MONTH_SEASON_YEAR").Object()
	// nature.ContainsKey("unit").ValueEqual("unit", "MONTH")
	// nature.ContainsKey("endInAdvance").ValueEqual("endInAdvance", 0)
	// rdtimes := bind.ContainsKey("REDEEM_TIMES").Value("REDEEM_TIMES").Object()
	// rdtimes.ContainsKey("times").ValueEqual("times", 1)
}

func __validateBasicCoupon(c *httpexpect.Object, cid string, cref string, channelID string, state int) {
	c.ContainsKey("ID").NotEmpty()
	c.ContainsKey("ConsumerID").ValueEqual("ConsumerID", cid)
	c.ContainsKey("ConsumerRefID").ValueEqual("ConsumerRefID", cref)
	c.ContainsKey("ChannelID").ValueEqual("ChannelID", channelID)
	c.ContainsKey("State").ValueEqual("State", state)
	c.ContainsKey("CreatedTime")
}

func _post1CouponWithToken(e *httpexpect.Expect, u string, uref string, t string, channelID string, statusCode int, token string) (*httpexpect.Object, string) {
	r := e.POST("/api/coupons/").WithHeader("Authorization", token).
		WithForm(map[string]interface{}{
			"consumerIDs":    u,
			"couponTypeID":   t,
			"consumerRefIDs": uref,
			"channelID":      channelID,
		}).Expect().Status(statusCode)

	var arr *httpexpect.Array
	if http.StatusOK == r.Raw().StatusCode {
		arr = r.JSON().Array()
	} else {
		return nil, ""
	}

	arr.Length().Equal(1)
	o := arr.Element(0).Object()
	return o, o.Value("ID").String().Raw()
}

func _post1Coupon(e *httpexpect.Expect, u string, uref string, t string, channelID string, statusCode int) (*httpexpect.Object, string) {
	return _post1CouponWithToken(e, u, uref, t, channelID, statusCode, jwtyhl5000d)
}

func _post2Coupons(e *httpexpect.Expect, u1 string, u2 string, u1ref string, u2ref string, t string, channelID string) (*httpexpect.Object, *httpexpect.Object, string, string) {
	arr := e.POST("/api/coupons/").WithHeader("Authorization", jwtyhl5000d).
		WithForm(map[string]interface{}{
			"consumerIDs":    strings.Join([]string{u1, u2}, ","),
			"couponTypeID":   t,
			"consumerRefIDs": strings.Join([]string{u1ref, u2ref}, ","),
			"channelID":      channelID,
		}).
		Expect().Status(http.StatusOK).
		JSON().Array()

	arr.Length().Equal(2)
	var c1, c2 *httpexpect.Object
	if arr.Element(0).Object().Value("ConsumerID").String().Raw() == u1 {
		c1 = arr.Element(0).Object()
		c2 = arr.Element(1).Object()
	} else {
		c1 = arr.Element(1).Object()
		c2 = arr.Element(0).Object()
	}
	return c1, c2, c1.Value("ID").String().Raw(), c2.Value("ID").String().Raw()
}

func _getCouponByID(e *httpexpect.Expect, cid string) *httpexpect.Object {
	return e.GET(strings.Join([]string{"/api/coupons/", cid}, "")).WithHeader("Authorization", jwtyhl5000d).
		Expect().Status(http.StatusOK).
		JSON().Object()
}

func _getCoupons(e *httpexpect.Expect, u string) *httpexpect.Array {
	arr := e.GET("/api/coupons/").WithHeader("Authorization", jwtyhl5000d).
		WithHeader("consumerID", u).
		Expect().Status(http.StatusOK).
		JSON().Array()
	return arr
}

func _redeemCouponByID(e *httpexpect.Expect, u string, cid string, extraInfo string, statusCode int) {
	e.POST("/api/redemptions").WithHeader("Authorization", jwtyhl5000d).
		WithForm(map[string]interface{}{
			"consumerIDs": u,
			"couponID":    cid,
			"extraInfo":   extraInfo,
		}).Expect().Status(statusCode)
	// v := r.JSON()
	// fmt.Println(v.String().Raw())
}

// func _prepareARandomCouponInDB(consumerID string, consumerRefID string, channelID string, couponType string, propties string) *Coupon {
// 	state := r.Intn(int(SUnknown))
// 	var p map[string]interface{}
// 	p = make(map[string]interface{}, 1)
// 	p["the_key"] = "the value"
// 	c := _aCoupon(consumerID, consumerRefID, channelID, couponType, State(state), p)
// 	sql := fmt.Sprintf(`INSERT INTO coupons VALUES ("%s","%s","%s","%s","%s",%d,"%s","%s")`, c.ID, c.CouponTypeID, c.ConsumerID, c.ConsumerRefID, c.ChannelID, c.State, propties, c.CreatedTime)
// 	_, _ = dbConnection.Exec(sql)
// 	return c
// }
