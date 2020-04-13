package main

import (
	"api-tests-for-coupon-service/base"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/google/uuid"
	"github.com/jinzhu/now"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_APPLY_TIMES(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("coupon A only can apply one time", t, func() {
		u := uuid.New().String()
		_post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeA, channelID, http.StatusOK)
		_post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeA, channelID, http.StatusBadRequest)
	})

	Convey("coupon B can apply upto 2 times", t, func() {
		u := uuid.New().String()
		_post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeB, channelID, http.StatusOK)
		_post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeB, channelID, http.StatusOK)
		_post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeB, channelID, http.StatusBadRequest)
	})
}

func Test_REDEEM_TIMES(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("coupon A only can redeem one time", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeA, channelID, http.StatusOK)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusOK)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusBadRequest)
	})

	Convey("coupon B can redeem upto 2 times", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeB, channelID, http.StatusOK)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusOK)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusOK)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusBadRequest)
	})
}

func Test_REDEEM_BY_SAME_BRAND(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("coupon issued by lancome can't be redeemed by parise", t, func() {
		u := uuid.New().String()
		_, cid := _post1CouponWithToken(e, u, strings.Join([]string{u, "xx"}, ""), typeB, channelID, http.StatusOK, jwttony5000d)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusBadRequest)
	})
}

func Test_REDEEM_IN_CURRENT_NATURE_MONTH_SEASON_YEAR(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("coupon A can not redeem in next month", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeA, channelID, http.StatusOK)
		ct := time.Now().AddDate(0, 0, -31)
		sql := fmt.Sprintf(`UPDATE coupons SET createdTime = "%s" WHERE id = "%s"`, ct, cid)
		_, _ = dbConnection.Exec(sql)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusBadRequest)
	})

	Convey("coupon C can not redeem since expired", t, func() {

		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeC, channelID, http.StatusOK)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusBadRequest)
	})

	Convey("coupon B can redeem in next month", t, func() {

		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeB, channelID, http.StatusOK)

		ct := now.BeginningOfMonth().AddDate(0, 0, -5).Local() // coupon b 延长了31天。
		sql := fmt.Sprintf(`UPDATE coupons SET createdTime = %d  WHERE id = "%s"`, ct.Unix(), cid)
		_, _ = dbConnection.Exec(sql)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusOK)
	})

	Convey("coupon D can not redeem in the month after next month", t, func() {

		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeD, channelID, http.StatusOK)

		ct := now.BeginningOfMonth().AddDate(0, 0, -35) // 因为月份天数不一致，如果今天是1号或者2号可能会失败。
		sql := fmt.Sprintf(`UPDATE coupons SET createdTime = %d  WHERE id = "%s"`, ct.Unix(), cid)
		_, _ = dbConnection.Exec(sql)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusBadRequest)
	})
}

func Test_REDEEM_PERIOD_WITH_OFFSET(t *testing.T) {
	e := httpexpect.New(t, baseurl)

	Convey("coupon E can redeem after issued", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeE, channelID, http.StatusOK)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusOK)
	})

	Convey("coupon E can redeem after 100 days", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeE, channelID, http.StatusOK)
		ct := time.Now().AddDate(0, 0, -100)
		sql := fmt.Sprintf(`UPDATE coupons SET createdTime = "%s" WHERE id = "%s"`, ct, cid)
		_, _ = dbConnection.Exec(sql)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusOK)
	})

	Convey("coupon E can not redeem after 365 days", t, func() {
		u := uuid.New().String()
		_, cid := _post1Coupon(e, u, strings.Join([]string{u, "xx"}, ""), typeE, channelID, http.StatusOK)
		ct := time.Now().AddDate(0, 0, -365)
		sql := fmt.Sprintf(`UPDATE coupons SET createdTime = "%s" WHERE id = "%s"`, ct, cid)
		_, _ = dbConnection.Exec(sql)
		_redeemCouponByID(e, u, cid, base.RandString(4), http.StatusOK)
	})
}
