package main

import (
	"api-tests-for-coupon-service/base"
	// "database/sql"
	// "fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"
)

var baseurl string = "http://127.0.0.1:1503"
var jwtyhl5000d string = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJvbXZRTEFoTWxVcUsxSjItQmhFYl82QlNkTFpmNkhSVVgtUlhXcHRINElFIn0.eyJqdGkiOiI5NWE1N2E0NS1lOGU3LTRhYzEtOTNhYi00NDI0M2ExNTg3YWMiLCJleHAiOjIwMTMyMjUwMTYsIm5iZiI6MCwiaWF0IjoxNTgxMjI1MDE2LCJpc3MiOiJodHRwOi8vNTIuMTMwLjczLjE4MC9hdXRoL3JlYWxtcy9Mb3JlYWwiLCJzdWIiOiIzNzk1MzdkYy1mYzdkLTRmMzAtOWExMy0wOWFjNmY4OGZhYjYiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJzb21lYXBwIiwiYXV0aF90aW1lIjowLCJzZXNzaW9uX3N0YXRlIjoiY2I2NDllOGQtZDkwNC00M2E3LWE2NDMtNGY1YTNmMDhmMmVkIiwiYWNyIjoiMSIsInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjb3Vwb25fcmVkZWVtZXIiLCJjb3Vwb25faXNzdWVyIl19LCJzY29wZSI6ImVtYWlsIHByb2ZpbGUiLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInByZWZlcnJlZF91c2VybmFtZSI6InlobDEwMDAwIiwiYnJhbmQiOiJMQU5DT01FIn0.jiRcTomtwvCnyFg2PxZA2MfK36UmD0tsx9kz3PNjlo4tNoxVPax-ocdln5qQfrJg6yzASwsg_-gNFgLhUqCCUV0iGLXcf69fBxvuxcQyBszmcByPda55u_zvLrv-91mI0a167ipjaIWqL2uOo_lSPm44JpwBcex2nqjz1FFBk1g3-nAHPuceh4-c0cF0y51QlS4fSCexorYDlIhUtcym6YQn9hmrM6Xdtjdtgf4iG_srnlH3gIAckT-Ihq_7rueNRE6cniabXg5AkzluEIwDwxY9KbPjSQ6Y1mxAleZ_dIvLFXzjxbnXn1vm8jRt3MAtvxG5yQ0sKjyb9j7h8jGzPA"
var jwttony5000d string = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJvbXZRTEFoTWxVcUsxSjItQmhFYl82QlNkTFpmNkhSVVgtUlhXcHRINElFIn0.eyJqdGkiOiJiYjUxNWYyZC01ZDA0LTQxYzctYWM0My00ZGRiMTQxODBiNDMiLCJleHAiOjIwMTQwNzg2OTAsIm5iZiI6MCwiaWF0IjoxNTgyMDc4NjkwLCJpc3MiOiJodHRwOi8vNTIuMTMwLjczLjE4MC9hdXRoL3JlYWxtcy9Mb3JlYWwiLCJzdWIiOiI5ZDU3ZGFmMi0xN2Q3LTRiYjMtYjk1Mi00ODY1MTMzNjkwMzgiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJzb21lYXBwIiwiYXV0aF90aW1lIjowLCJzZXNzaW9uX3N0YXRlIjoiZjAwNDhhMWItOGFjOC00ZTA1LWI2Y2QtN2RhZTNmYTU4ZjFhIiwiYWNyIjoiMSIsInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjb3Vwb25fcmVkZWVtZXIiLCJjb3Vwb25faXNzdWVyIl19LCJzY29wZSI6ImVtYWlsIHByb2ZpbGUiLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInByZWZlcnJlZF91c2VybmFtZSI6InRvbnkiLCJicmFuZCI6IlBBUklTX0xPUkVBTCJ9.Z0-ZKZXcqcATnjDpfVYFelq4scCBMP1l9LuCC8zUYmzJwBgo56JwdhVaQO-_sQeaqU__a7gGz8P2DKxv_Y6mbc5jTh5BWcc7AC9LR5axZFzHVTgp_ZE7FCBEHkmYpF72W6BKI73e-0_Qwn1FVRxWAF7KkuSnV5_hdfi6-CStfREikE5_Wr0VGS9mn_fcmuVbGchE1yzHhDGKmVa2RiypcMWGHcSY6iF9FTkbF9TbZ-Lu-ASJRFQ8U-k7Q4wtd8laMQTctEJHMVXQ0GcQ362J5l42OiCZjjTleMghx05gvbjhaCF8FI0YdgaBkPUQ-hw_C4IO2fJdc6x4CkdTolgHQg"
var jwt1s string = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJvbXZRTEFoTWxVcUsxSjItQmhFYl82QlNkTFpmNkhSVVgtUlhXcHRINElFIn0.eyJqdGkiOiJmZWExMDE3NS1jM2Q4LTQ5ZDEtYjI1NS1kODcwYTJiNWFmYmQiLCJleHAiOjE1ODEyMjUxNjgsIm5iZiI6MCwiaWF0IjoxNTgxMjI1MTA4LCJpc3MiOiJodHRwOi8vNTIuMTMwLjczLjE4MC9hdXRoL3JlYWxtcy9Mb3JlYWwiLCJzdWIiOiIzNzk1MzdkYy1mYzdkLTRmMzAtOWExMy0wOWFjNmY4OGZhYjYiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJzb21lYXBwIiwiYXV0aF90aW1lIjowLCJzZXNzaW9uX3N0YXRlIjoiYjJkODdmMjYtMGEyNC00MTQxLWI1MmItMWJhNTFiOGQ4NmYxIiwiYWNyIjoiMSIsInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjb3Vwb25fcmVkZWVtZXIiLCJjb3Vwb25faXNzdWVyIl19LCJzY29wZSI6ImVtYWlsIHByb2ZpbGUiLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInByZWZlcnJlZF91c2VybmFtZSI6InlobDEwMDAwIiwiYnJhbmQiOiJMQU5DT01FIn0.wACD6pdlnGEZdgMmToCQnp29L5wTxjwsCBB0qPNhULZ3ZSIXAIoC7bb3Rjzomk_FpJjCpBlcmfm83kU1UbDQuiKSkI6DOemZMbccAfHfnqEu3V7195-pBTIWsEpVpKNCFI4lXAKBo7IsHBJwWfMrdkSdUljYWIC_7LgH3vVmY6LheEszRcl9P3NbXaDcmWYtjuywIS7Aph5wse8671aJ7w2ahyDLsr7prlUNs0K9rJMgGy1DNJhzVaGXQnEmg2IMWQihJT1YGzWTzTE24YieQ0BYzvHPfwPsDxyiDZS6qj3z9qBugmFAgJulU7AnoAmEdEFSdMHN_0QwqlvzhGLlrg"
var jwtnoissueredeemrole string = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJvbXZRTEFoTWxVcUsxSjItQmhFYl82QlNkTFpmNkhSVVgtUlhXcHRINElFIn0.eyJqdGkiOiI5YTUxNDU0My02YWQ3LTRmMjMtOGJiYS1lZGFmNzI5NDRlZDEiLCJleHAiOjIwMTMzMDQ4MTYsIm5iZiI6MCwiaWF0IjoxNTgxMzA0ODE2LCJpc3MiOiJodHRwOi8vNTIuMTMwLjczLjE4MC9hdXRoL3JlYWxtcy9Mb3JlYWwiLCJzdWIiOiIzNzk1MzdkYy1mYzdkLTRmMzAtOWExMy0wOWFjNmY4OGZhYjYiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJzb21lYXBwIiwiYXV0aF90aW1lIjowLCJzZXNzaW9uX3N0YXRlIjoiOGU2NzVhYTUtZDMxZi00YTFkLWI3OGItYjllYjBhOTYyODg1IiwiYWNyIjoiMSIsInNjb3BlIjoiZW1haWwgcHJvZmlsZSIsImVtYWlsX3ZlcmlmaWVkIjpmYWxzZSwicHJlZmVycmVkX3VzZXJuYW1lIjoieWhsMTAwMDAiLCJicmFuZCI6IkxBTkNPTUUifQ.PwCTxuLcMwz4-XyoBd750rk7T2qXU6zBdLtSEUa7pMGPI5Eo7Ayp0Yub15EraIoZP0Kphv9MdFtKkNMuV4Ua9otSDkPe2CSn4be3Ez-gvhk7Gbylh1atyrSpdaW-QJNqCek3C8jvGnC4c3_9o4Bduj6pnrRtrttM-oEFGrtauahIr73vmBuRalokw7OMm2dfq3ot8hTb1oT5RkPt_IILxTIorxMKUSoetKUM9b87KHv7EMZQz0sZne8EQ6DrpZmZDAsxU4RL5osKpgYH6p7XACG8RznZtDDfN0uC87nRiUyRbHYHetRwyu_AlnxpRfyCtFCrOFYn00hdrQYO7wi0FA"
var typeA string = "63f9f1ce-2ad0-462a-b798-4ead5e5ab3a5"
var typeB string = "58b388ff-689e-445a-8bbd-8d707dbe70ef"
var typeC string = "abd73dbe-cc91-4b61-b10c-c6532d7a7770"
var typeD string = "ca0ff68f-dc05-488d-b185-660b101a1068"
var typeE string = "7c0fb6b2-7362-4933-ad15-cd4ad9fccfec"
var channelID string = "defaultChannel"

// TODO: 需要直接在DB插入数据然后做一些验证，因为现在无法测试一些和时间有关的case

func Test_Authorization(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("Given a request without Authorization", t, func() {
		e.GET("/api/coupontypes").
			Expect().Status(http.StatusUnauthorized)
	})

	Convey("Given a request with a expired jwt", t, func() {
		e.GET("/api/coupontypes").
			WithHeader("Authorization", jwt1s).
			Expect().Status(http.StatusBadRequest).
			JSON().Object().ContainsKey("error-code").ValueEqual("error-code", 1501)
	})
}

func Test_GET_coupontypes(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("Make sure api only support GET", t, func() {
		e.POST("/api/coupontypes").WithHeader("Authorization", jwtyhl5000d).
			Expect().Status(http.StatusMethodNotAllowed)
		e.PUT("/api/coupontypes").WithHeader("Authorization", jwtyhl5000d).
			Expect().Status(http.StatusMethodNotAllowed)
		e.DELETE("/api/coupontypes").WithHeader("Authorization", jwtyhl5000d).
			Expect().Status(http.StatusMethodNotAllowed)
		e.PATCH("/api/coupontypes").WithHeader("Authorization", jwtyhl5000d).
			Expect().Status(http.StatusMethodNotAllowed)
	})

	Convey("Given a person with no required roles", t, func() {
		e.GET("/api/coupontypes").
			WithHeader("Authorization", jwtnoissueredeemrole).
			Expect().Status(http.StatusForbidden)
	})

	Convey("Given a right person to get the coupon types", t, func() {
		arr := e.GET("/api/coupontypes").
			WithHeader("Authorization", jwtyhl5000d).
			Expect().Status(http.StatusOK).
			JSON().Array()
		arr.Length().Gt(0)
		obj := arr.Element(0).Object()
		obj.ContainsKey("name")
		obj.ContainsKey("id")
		obj.ContainsKey("template_id")
		obj.ContainsKey("description")
		obj.ContainsKey("internal_description")
		obj.ContainsKey("state")
		obj.ContainsKey("publisher")
		obj.ContainsKey("visible_start_time")
		obj.ContainsKey("visible_end_time")
		obj.ContainsKey("created_time")
		obj.ContainsKey("deleted_time")
		obj.ContainsKey("rules")
	})
}

func Test_POST_coupons(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("Issue coupon A for one consumer", t, func() {
		u := uuid.New().String()
		c, _ := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeA, channelID, http.StatusOK)
		_validateCouponWithTypeA(c, u, strings.Join([]string{u, "xx"}, ""), channelID, 0)
	})

	Convey("Issue coupon of type A for multi consumers", t, func() {
		u1 := uuid.New().String()
		u2 := uuid.New().String()
		c1, c2, _, _ := _post2Coupons(e, u1, u2, strings.Join([]string{u1, "xx"}, ""), strings.Join([]string{u2, "yy"}, ""), typeA, channelID)

		Convey("the coupons should be same as created", func() {
			_validateCouponWithTypeA(c1, u1, strings.Join([]string{u1, "xx"}, ""), channelID, 0)
			_validateCouponWithTypeA(c2, u2, strings.Join([]string{u2, "yy"}, ""), channelID, 0)
		})

	})
}

func Test_Get_coupon(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("Issue coupon A for one consumer", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeA, channelID, http.StatusOK)
		// cid := oc.Value("ID").String().Raw()

		Convey("Get the coupon and validate", func() {
			c := _getCouponByID(e, cid)
			_validateCouponWithTypeA(c, u, strings.Join([]string{u, "xx"}, ""), channelID, 0)
		})
	})

	Convey("Issue coupon of type A for multi consumers", t, func() {
		u1 := uuid.New().String()
		u2 := uuid.New().String()
		_, _, cid1, cid2 := _post2Coupons(e, u1, u2, strings.Join([]string{u1, "xx"}, ""), strings.Join([]string{u2, "yy"}, ""), typeA, channelID)

		Convey("Get the coupon c1", func() {
			c1 := _getCouponByID(e, cid1)
			_validateCouponWithTypeA(c1, u1, strings.Join([]string{u1, "xx"}, ""), channelID, 0)
		})
		Convey("Get the coupon c2", func() {
			c2 := _getCouponByID(e, cid2)
			_validateCouponWithTypeA(c2, u2, strings.Join([]string{u2, "yy"}, ""), channelID, 0)
		})
	})
}

func Test_Get_coupons(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("Issue coupon A, B for one consumer", t, func() {
		u := uuid.New().String()
		uref := strings.Join([]string{u, "xx"}, "")
		_, cid1 := _post1Coupon(e, u, uref, typeA, channelID, http.StatusOK)
		_post1Coupon(e, u, uref, typeB, channelID, http.StatusOK)
		// _, _, cid1, _ := _post2Coupons(e, u, u, uref, uref, typeA, typeB, channelID)

		Convey("Get the coupons", func() {
			arr := _getCoupons(e, u)
			arr.Length().Equal(2)
			var c1, c2 *httpexpect.Object
			if arr.Element(0).Object().Value("ID").String().Raw() == cid1 {
				c1 = arr.Element(0).Object()
				c2 = arr.Element(1).Object()
			} else {
				c1 = arr.Element(1).Object()
				c2 = arr.Element(0).Object()
			}

			Convey("the coupons should be same as created", func() {
				_validateCouponWithTypeA(c1, u, strings.Join([]string{u, "xx"}, ""), channelID, 0)
				_validateCouponWithTypeB(c2, u, strings.Join([]string{u, "xx"}, ""), channelID, 0)
			})
		})
	})
}

func Test_POST_redemptions(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("Given a coupon for redeem by coupon id", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeA, channelID, http.StatusOK)

		Convey("Redeem the coupon", func() {
			_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusOK)
			_getCouponByID(e, cid).ContainsKey("State").ValueEqual("State", 2)
		})
	})

	Convey("Given 2 coupons to redeem by coupon type", t, func() {
		u1 := uuid.New().String()
		u2 := uuid.New().String()
		_, _, cid1, cid2 := _post2Coupons(e, u1, u2, strings.Join([]string{u1, "xx"}, ""), strings.Join([]string{u2, "yy"}, ""), typeA, channelID)

		Convey("Redeem the coupons", func() {
			extraInfo := base.RandString(4)
			e.POST("/api/redemptions").WithHeader("Authorization", jwtyhl5000d).
				WithForm(map[string]interface{}{
					"consumerIDs": strings.Join([]string{u1, u2}, ","),
					// "couponID":    cid,
					"couponTypeID": typeA,
					"extraInfo":    extraInfo,
				}).Expect().Status(http.StatusOK)

			_getCouponByID(e, cid1).ContainsKey("State").ValueEqual("State", 2)
			_getCouponByID(e, cid2).ContainsKey("State").ValueEqual("State", 2)
		})
	})

	Convey("Redeem coupon should bind correct consumer", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeA, channelID, http.StatusOK)

		Convey("Redeem the coupon by another one should failed", func() {
			_redeemCouponByID(e, "hacker", cid, base.RandString(4), http.StatusBadRequest)
			_getCouponByID(e, cid).ContainsKey("State").ValueEqual("State", 0)
		})
	})
}
