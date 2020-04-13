package coupon

import (
	"database/sql"
	// "errors"
	// "log"
	// "fmt"
	// "net/http"
	"reflect"
	// "strings"
	"testing"
	// "time"

	"loreal.com/dit/cmd/coupon-service/base"

	// "github.com/google/uuid"
	"bou.ke/monkey"
	. "github.com/smartystreets/goconvey/convey"
)

// var r *rand.Rand = rand.New(rand.NewSource(time.Now().Unix()))

func Test_IssueCoupons(t *testing.T) {
	Convey("Given a request without issue role", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		// consumerID := base.RandString(4)
		// couponTypeID := uuid.New().String()
		// transLog := base.RandString(40)
		// data := base.RandString(30)
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return false
		})
		Convey("Should be ErrRequesterForbidden", func() {
			_, _, err := IssueCoupons(requester, "", "", "", "")
			So(err, ShouldEqual, &ErrRequesterForbidden)
		})
		pg.Unpatch()
	})

	Convey("Given a non existed coupon type", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return nil
		})
		Convey("Should be ErrCouponTemplateNotFound", func() {
			_, _, err := IssueCoupons(requester, "", "", "", "")
			So(err, ShouldEqual, &ErrCouponTemplateNotFound)
		})
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that ErrConsumerIDsAndRefIDsMismatch", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		Convey("Given ref ids less than consumer ids", func() {
			_, _, err := IssueCoupons(requester, "abc,def", "111", "", "")
			Convey("Should be ErrRuleNotFound", func() {
				So(err, ShouldEqual, &ErrConsumerIDsAndRefIDsMismatch)
			})
		})

		Convey("Given ref ids more than consumer ids", func() {
			_, _, err := IssueCoupons(requester, "abc,def", "111,222,333", "", "")
			Convey("Should be ErrRuleNotFound", func() {
				So(err, ShouldEqual, &ErrConsumerIDsAndRefIDsMismatch)
			})
		})

		Convey("Given ref ids more than consumer ids and one ref id is blank", func() {
			_, _, err := IssueCoupons(requester, "abc,def", "111,,333", "", "")
			Convey("Should be ErrRuleNotFound", func() {
				So(err, ShouldEqual, &ErrConsumerIDsAndRefIDsMismatch)
			})
		})

		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that validateTemplateRules will throw an err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(_ string, _ string, _ *PublishedCouponType) ([]error, error) {
			return nil, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, _, err := IssueCoupons(requester, "abc", "", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that validateTemplateRules will throw multi errs", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(cid string, _ string, _ *PublishedCouponType) ([]error, error) {
			switch cid {
			case "google":
				return []error{&ErrRuleNotFound, &ErrCouponTemplateNotFound}, nil
			case "apple":
				return []error{&ErrRuleNotFound, &ErrCouponTemplateNotFound, &ErrRequesterForbidden}, nil
			}
			return nil, nil
		})

		Convey("Should be many errors", func() {
			_, errs, _ := IssueCoupons(requester, "google,apple", "", "", "")
			So(errs, ShouldNotBeNil)
			So(len(errs), ShouldEqual, 2)
			gerrs := errs["google"]
			aerrs := errs["apple"]
			So(len(gerrs), ShouldEqual, 2)
			So(gerrs[0], ShouldEqual, &ErrRuleNotFound)
			So(gerrs[1], ShouldEqual, &ErrCouponTemplateNotFound)
			So(len(aerrs), ShouldEqual, 3)
			So(aerrs[0], ShouldEqual, &ErrRuleNotFound)
			So(aerrs[1], ShouldEqual, &ErrCouponTemplateNotFound)
			So(aerrs[2], ShouldEqual, &ErrRequesterForbidden)
		})
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that marshalCouponRules throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(_ string, _ string, _ *PublishedCouponType) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(marshalCouponRules, func(_ *base.Requester, _ string, _ map[string]map[string]interface{}) (map[string]interface{}, error) {
			return nil, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, _, err := IssueCoupons(requester, "google,apple", "", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that transaction throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(_ string, _ string, _ *PublishedCouponType) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(marshalCouponRules, func(_ *base.Requester, _ string, _ map[string]map[string]interface{}) (map[string]interface{}, error) {
			return nil, nil
		})

		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return nil, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, _, err := IssueCoupons(requester, "google", "", "", "")
			So(err, ShouldNotBeNil)
		})
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that createCoupons throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(_ string, _ string, _ *PublishedCouponType) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(marshalCouponRules, func(_ *base.Requester, _ string, _ map[string]map[string]interface{}) (map[string]interface{}, error) {
			return nil, nil
		})

		mockSQLTx := new(sql.Tx)

		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCoupons, func(_ []*Coupon) error {
			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, _, err := IssueCoupons(requester, "google", "", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Check createCoupons get the right data", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(_ string, _ string, _ *PublishedCouponType) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(marshalCouponRules, func(_ *base.Requester, _ string, _ map[string]map[string]interface{}) (map[string]interface{}, error) {
			rules := map[string]interface{}{
				"abc": "def",
				"xyz": "123",
			}
			return rules, nil
		})

		mockSQLTx := new(sql.Tx)

		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		//这里用来判断方法内部得到的数据是否是期望的数据，类似mokito里面的captor
		pg7 := monkey.Patch(createCoupons, func(cs []*Coupon) error {
			if len(cs) != 2 {
				return &ErrCouponTemplateNotFound
			}
			if cs[0].CouponTypeID != "tttt" {
				return &ErrCouponTemplateNotFound
			}
			if cs[0].ConsumerID != "google" && cs[1].ConsumerID != "apple" {
				return &ErrCouponTemplateNotFound
			}
			if cs[0].State != SActive && cs[1].State != SActive {
				return &ErrCouponTemplateNotFound
			}
			rules := cs[0].Properties[KeyBindingRuleProperties].(map[string]interface{})
			if len(rules) != 2 || rules["abc"] != "def" || rules["xyz"] != "123" {
				return &ErrCouponTemplateNotFound
			}

			return &ErrRuleNotFound // 这个错误表示得到想要的输入参数
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, _, err := IssueCoupons(requester, "google,apple", "", "", "tttt")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that createCouponTransactions throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(_ string, _ string, _ *PublishedCouponType) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(marshalCouponRules, func(_ *base.Requester, _ string, _ map[string]map[string]interface{}) (map[string]interface{}, error) {
			rules := map[string]interface{}{
				"abc": "def",
				"xyz": "123",
			}
			return rules, nil
		})

		mockSQLTx := new(sql.Tx)

		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCoupons, func(cs []*Coupon) error {
			return nil
		})

		pg8 := monkey.Patch(createCouponTransactions, func(ts []*Transaction) error {
			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, _, err := IssueCoupons(requester, "google", "", "", "tttt")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Check createCouponTransactions get the right data", t, func() {
		rid := base.RandString(4)
		var requester = _aRequester(rid, []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(_ string, _ string, _ *PublishedCouponType) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(marshalCouponRules, func(_ *base.Requester, _ string, _ map[string]map[string]interface{}) (map[string]interface{}, error) {
			rules := map[string]interface{}{
				"abc": "def",
				"xyz": "123",
			}
			return rules, nil
		})

		mockSQLTx := new(sql.Tx)

		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCoupons, func(cs []*Coupon) error {
			return nil
		})

		pg8 := monkey.Patch(createCouponTransactions, func(ts []*Transaction) error {
			if len(ts) != 2 {
				return &ErrCouponTemplateNotFound
			}
			if ts[0].ActorID != rid && ts[1].ActorID != rid {
				return &ErrCouponTemplateNotFound
			}
			//TODO: 校验couponid是否一致
			if ts[0].TransType != TTIssueCoupon && ts[1].TransType != TTIssueCoupon {
				return &ErrCouponTemplateNotFound
			}

			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, _, err := IssueCoupons(requester, "google,apple", "", "", "tttt")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that everything is Okay", t, func() {
		rid := base.RandString(4)
		var requester = _aRequester(rid, []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(_checkCouponType, func(_ string) *PublishedCouponType {
			return new(PublishedCouponType)
		})

		pg3 := monkey.Patch(validateTemplateRules, func(_ string, _ string, _ *PublishedCouponType) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(marshalCouponRules, func(_ *base.Requester, _ string, _ map[string]map[string]interface{}) (map[string]interface{}, error) {
			rules := map[string]interface{}{
				"abc": "def",
				"xyz": "123",
			}
			return rules, nil
		})

		mockSQLTx := new(sql.Tx)

		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCoupons, func(cs []*Coupon) error {
			return nil
		})

		pg8 := monkey.Patch(createCouponTransactions, func(ts []*Transaction) error {
			return nil
		})

		pg9 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Commit", func(*sql.Tx) error {
			return nil
		})

		Convey("Should get the new coupons", func() {
			cs, _, _ := IssueCoupons(requester, "google,apple", "", "", "tttt")
			So(len(*cs), ShouldEqual, 2)
			So((*cs)[0].CouponTypeID, ShouldEqual, "tttt")
			So((*cs)[0].ConsumerID, ShouldBeIn, "google", "apple")
			So((*cs)[1].ConsumerID, ShouldBeIn, "google", "apple")
			So((*cs)[0].State, ShouldEqual, SActive)
			So((*cs)[1].State, ShouldEqual, SActive)
			rules := (*cs)[0].Properties[KeyBindingRuleProperties].(map[string]interface{})
			So(len(rules), ShouldEqual, 2)
			So(rules["abc"], ShouldEqual, "def")
			So(rules["xyz"], ShouldEqual, "123")
		})
		pg9.Unpatch()
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})
}

func Test_GetCoupon_CS(t *testing.T) {
	Convey("Given a request without issue or redeem role", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return false
		})
		Convey("Should be ErrRequesterForbidden", func() {
			_, err := GetCoupon(requester, "")
			So(err, ShouldEqual, &ErrRequesterForbidden)
		})
		pg.Unpatch()
	})

	Convey("Given a blank coupon id", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(base.IsBlankString, func(_ string) bool {
			return true
		})
		Convey("Should be ErrCouponIDInvalid", func() {
			_, err := GetCoupon(requester, "")
			So(err, ShouldEqual, &ErrCouponIDInvalid)
		})
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that getCoupon from db throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(base.IsBlankString, func(_ string) bool {
			return false
		})

		pg3 := monkey.Patch(getCoupon, func(_ string) (*Coupon, error) {
			return nil, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := GetCoupon(requester, "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that everything is Okay", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(base.IsBlankString, func(_ string) bool {
			return false
		})
		cid := "abc"
		rid := "xxx"
		chid := "yyy"
		tid := "def"
		st := SActive
		c := _aCoupon(cid, rid, chid, tid, st, nil)
		pg3 := monkey.Patch(getCoupon, func(_ string) (*Coupon, error) {
			return c, nil
		})

		Convey("Should get the right coupon", func() {
			c, _ := GetCoupon(requester, "")
			So(c, ShouldNotBeNil)
			So(c.ConsumerID, ShouldEqual, cid)
			So(c.ConsumerRefID, ShouldEqual, rid)
			So(c.ChannelID, ShouldEqual, chid)
			So(c.CouponTypeID, ShouldEqual, tid)
			So(c.State, ShouldEqual, st)
		})
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})
}

func Test_GetCoupons_CS(t *testing.T) {
	Convey("Given a request without issue or redeem role", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return false
		})
		Convey("Should be ErrRequesterForbidden", func() {
			_, err := GetCoupons(requester, "", "")
			So(err, ShouldEqual, &ErrRequesterForbidden)
		})
		pg.Unpatch()
	})

	Convey("Given a blank consumer id", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(base.IsBlankString, func(_ string) bool {
			return true
		})
		Convey("Should be ErrConsumerIDInvalid", func() {
			_, err := GetCoupons(requester, "", "")
			So(err, ShouldEqual, &ErrConsumerIDInvalid)
		})
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that getCoupons from db throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(base.IsBlankString, func(_ string) bool {
			return false
		})

		pg3 := monkey.Patch(getCoupons, func(_ string, _ string) ([]*Coupon, error) {
			return nil, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := GetCoupons(requester, "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that everything is Okay", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(base.IsBlankString, func(_ string) bool {
			return false
		})
		cs := make([]*Coupon, 2)
		pg3 := monkey.Patch(getCoupons, func(_ string, _ string) ([]*Coupon, error) {
			return cs, nil
		})

		Convey("Should get the right coupon", func() {
			cs, _ := GetCoupons(requester, "", "")
			So(len(cs), ShouldEqual, 2)
		})
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})
}

func Test_RedeemCoupon(t *testing.T) {
	Convey("Given a request without redeem role", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return false
		})
		Convey("Should be ErrRequesterForbidden", func() {
			_, err := RedeemCoupon(requester, "", "", "")
			So(err, ShouldEqual, &ErrRequesterForbidden)
		})
		pg.Unpatch()
	})

	Convey("Given an env that GetCoupon throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return nil, &ErrCouponTemplateNotFound
		})
		Convey("Should be ErrCouponTemplateNotFound", func() {
			_, err := RedeemCoupon(requester, "", "", "")
			So(err, ShouldEqual, &ErrCouponTemplateNotFound)
		})
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that validateCouponBasicForRedeem will throw an err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return nil, nil
		})

		pg3 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCoupon(requester, "", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that validateCouponRules will throw an err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return nil, nil
		})
		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCoupon(requester, "", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that validateCouponRules will throw multi errs", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return nil, nil
		})
		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return []error{&ErrRuleNotFound, &ErrCouponTemplateNotFound}, nil
		})

		Convey("Should be many errors", func() {
			errs, _ := RedeemCoupon(requester, "", "", "")
			So(errs, ShouldNotBeNil)
			So(len(errs), ShouldEqual, 2)
			So(errs[0], ShouldEqual, &ErrRuleNotFound)
			So(errs[1], ShouldEqual, &ErrCouponTemplateNotFound)
		})

		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that transaction throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return nil, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return nil, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCoupon(requester, "", "", "")
			So(err, ShouldNotBeNil)
		})
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that createCouponTransaction throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return nil, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		mockSQLTx := new(sql.Tx)
		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCouponTransaction, func(_ *Transaction) error {
			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCoupon(requester, "", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given restRedeemTimes throw an err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return nil, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		mockSQLTx := new(sql.Tx)
		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCouponTransaction, func(_ *Transaction) error {
			return nil
		})

		pg8 := monkey.Patch(restRedeemTimes, func(_ *Coupon) (uint, error) {
			return 0, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCoupon(requester, "", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given restRedeemTimes throw ErrCouponRulesNoRedeemTimes", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})

		c := Coupon{}
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return &c, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		mockSQLTx := new(sql.Tx)
		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCouponTransaction, func(_ *Transaction) error {
			return nil
		})

		pg8 := monkey.Patch(restRedeemTimes, func(_ *Coupon) (uint, error) {
			return 0, &ErrCouponRulesNoRedeemTimes
		})

		pg9 := monkey.Patch(updateCouponState, func(_ *Coupon) error {
			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCoupon(requester, "", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg9.Unpatch()
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given restRedeemTimes great than 0", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupon, func(_ *base.Requester, _ string) (*Coupon, error) {
			return nil, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		mockSQLTx := new(sql.Tx)
		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCouponTransaction, func(_ *Transaction) error {
			return nil
		})

		pg8 := monkey.Patch(restRedeemTimes, func(_ *Coupon) (uint, error) {
			return 5, nil
		})

		pg9 := monkey.Patch(updateCouponState, func(_ *Coupon) error {
			return &ErrRuleNotFound
		})
		bCommited := false
		pg10 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Commit", func(*sql.Tx) error {
			bCommited = true
			return nil
		})
		bSendMsg := false
		pg11 := monkey.Patch(_sendRedeemMessage, func([]*Coupon, string) {
			bSendMsg = true
		})

		Convey("Should commited and sent msg", func() {
			_, _ = RedeemCoupon(requester, "", "", "")
			So(bCommited, ShouldBeTrue)
			So(bSendMsg, ShouldBeTrue)
		})

		pg11.Unpatch()
		pg10.Unpatch()
		pg9.Unpatch()
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})
}

func Test_RedeemCouponByType(t *testing.T) {
	Convey("Given a request without redeem role", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return false
		})
		Convey("Should be ErrRequesterForbidden", func() {
			_, err := RedeemCouponByType(nil, requester, "", "", "")
			So(err, ShouldEqual, &ErrRequesterForbidden)
		})
		pg.Unpatch()
	})

	Convey("Given an env that GetCoupons throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			return nil, &ErrCouponTemplateNotFound
		})
		Convey("Should be ErrCouponTemplateNotFound", func() {
			errMap, _ := RedeemCouponByType(nil, requester, "abc", "", "")
			So(len(errMap), ShouldEqual, 1)
			So(len(errMap["abc"]), ShouldEqual, 1)
			So(errMap["abc"][0], ShouldEqual, &ErrCouponTemplateNotFound)
		})
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that GetCoupons get wrong data", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})

		Convey("If nil, should be ErrRuleNotFound", func() {
			pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
				return nil, nil
			})
			errMap, _ := RedeemCouponByType(nil, requester, "abc", "", "")
			So(len(errMap), ShouldEqual, 1)
			So(len(errMap["abc"]), ShouldEqual, 1)
			So(errMap["abc"][0], ShouldEqual, &ErrCouponNotFound)
			pg2.Unpatch()
		})

		Convey("If empty, should be ErrRuleNotFound", func() {
			pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
				return []*Coupon{}, nil
			})
			errMap, _ := RedeemCouponByType(nil, requester, "abc", "", "")
			So(len(errMap), ShouldEqual, 1)
			So(len(errMap["abc"]), ShouldEqual, 1)
			So(errMap["abc"][0], ShouldEqual, &ErrCouponNotFound)
			pg2.Unpatch()
		})

		Convey("If more than one coupon, should be ErrCouponTooMuchToRedeem", func() {
			pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
				return make([]*Coupon, 2), nil
			})
			errMap, _ := RedeemCouponByType(nil, requester, "abc", "", "")
			So(len(errMap), ShouldEqual, 1)
			So(len(errMap["abc"]), ShouldEqual, 1)
			So(errMap["abc"][0], ShouldEqual, &ErrCouponTooMuchToRedeem)
			pg2.Unpatch()
		})

		pg.Unpatch()
	})

	Convey("Given an env that validateCouponBasicForRedeem will throw an err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
			return []*Coupon{c}, nil
		})

		pg3 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			errMap, _ := RedeemCouponByType(nil, requester, "abc", "", "")
			So(len(errMap), ShouldEqual, 1)
			So(len(errMap["abc"]), ShouldEqual, 1)
			So(errMap["abc"][0], ShouldEqual, &ErrRuleNotFound)
		})
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that validateCouponRules will throw an err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
			return []*Coupon{c}, nil
		})
		pg3 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		Convey("for the case only one consumer", func() {
			pg4 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
				return nil, &ErrRuleNotFound
			})

			Convey("Should be ErrRuleNotFound", func() {
				errMap, _ := RedeemCouponByType(nil, requester, "abc", "", "")
				So(len(errMap), ShouldEqual, 1)
				So(len(errMap["abc"]), ShouldEqual, 1)
				So(errMap["abc"][0], ShouldEqual, &ErrRuleNotFound)
			})
			pg4.Unpatch()
		})

		Convey("for the case with 2 consumers", func() {
			Convey("both consumers have errors", func() {
				pg4 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
					return nil, &ErrRuleNotFound
				})
				pg5 := monkey.Patch(GetCoupons, func(_ *base.Requester, consumerID string, _ string) ([]*Coupon, error) {
					c := _aCoupon(consumerID, "xxx", "yyy", "def", 0, nil)
					return []*Coupon{c}, nil
				})
				Convey("Should be ErrRuleNotFound", func() {
					errMap, _ := RedeemCouponByType(nil, requester, "abc,def", "", "")
					So(len(errMap), ShouldEqual, 2)
					So(len(errMap["abc"]), ShouldEqual, 1)
					So(len(errMap["def"]), ShouldEqual, 1)
					So(errMap["abc"][0], ShouldEqual, &ErrRuleNotFound)
					So(errMap["def"][0], ShouldEqual, &ErrRuleNotFound)
				})
				pg4.Unpatch()
				pg5.Unpatch()
			})

			Convey("first consumer has error", func() {
				pg4 := monkey.Patch(validateCouponRules, func(_ *base.Requester, consumerID string, _ *Coupon) ([]error, error) {
					if consumerID == "abc" {
						return nil, &ErrRuleNotFound
					}
					return nil, nil
				})
				pg5 := monkey.Patch(GetCoupons, func(_ *base.Requester, consumerID string, _ string) ([]*Coupon, error) {
					c := _aCoupon(consumerID, "xxx", "yyy", "def", 0, nil)
					return []*Coupon{c}, nil
				})

				Convey("Should be ErrRuleNotFound", func() {
					errMap, _ := RedeemCouponByType(nil, requester, "abc,def", "", "")
					So(len(errMap), ShouldEqual, 1)
					So(len(errMap["abc"]), ShouldEqual, 1)
					So(errMap["abc"][0], ShouldEqual, &ErrRuleNotFound)
				})
				pg4.Unpatch()
				pg5.Unpatch()
			})

			Convey("second consumer has error", func() {
				pg4 := monkey.Patch(validateCouponRules, func(_ *base.Requester, consumerID string, _ *Coupon) ([]error, error) {
					if consumerID == "def" {
						return nil, &ErrRuleNotFound
					}
					return nil, nil
				})
				pg5 := monkey.Patch(GetCoupons, func(_ *base.Requester, consumerID string, _ string) ([]*Coupon, error) {
					c := _aCoupon(consumerID, "xxx", "yyy", "def", 0, nil)
					return []*Coupon{c}, nil
				})

				Convey("Should be ErrRuleNotFound", func() {
					errMap, _ := RedeemCouponByType(nil, requester, "abc,def", "", "")
					So(len(errMap), ShouldEqual, 1)
					So(len(errMap["def"]), ShouldEqual, 1)
					So(errMap["def"][0], ShouldEqual, &ErrRuleNotFound)
				})
				pg4.Unpatch()
				pg5.Unpatch()
			})
		})

		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that validateCouponRules will throw multi errs", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
			return []*Coupon{c}, nil
		})
		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return []error{&ErrRuleNotFound, &ErrCouponTemplateNotFound}, nil
		})

		Convey("Should be many errors", func() {
			errMap, _ := RedeemCouponByType(nil, requester, "abc", "", "")
			So(errMap, ShouldNotBeNil)
			So(len(errMap["abc"]), ShouldEqual, 2)
			So(errMap["abc"][0], ShouldEqual, &ErrRuleNotFound)
			So(errMap["abc"][1], ShouldEqual, &ErrCouponTemplateNotFound)
		})

		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that transaction throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
			return []*Coupon{c}, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return nil, &ErrRuleNotFound
		})

		pg6 := monkey.Patch(_sendRedeemMessage, func(_ []*Coupon, _ string) {
			return
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCouponByType(nil, requester, "abc", "", "")
			So(err, ShouldNotBeNil)
		})
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given an env that createCouponTransaction throw err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
			return []*Coupon{c}, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		mockSQLTx := new(sql.Tx)
		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCouponTransactions, func(_ []*Transaction) error {
			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCouponByType(nil, requester, "abc", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given restRedeemTimes throw an err", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
			return []*Coupon{c}, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		mockSQLTx := new(sql.Tx)
		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCouponTransactions, func(_ []*Transaction) error {
			return nil
		})

		pg8 := monkey.Patch(restRedeemTimes, func(_ *Coupon) (uint, error) {
			return 0, &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCouponByType(nil, requester, "abc", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given restRedeemTimes throw ErrCouponRulesNoRedeemTimes", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})

		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
			return []*Coupon{c}, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		mockSQLTx := new(sql.Tx)
		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCouponTransactions, func(_ []*Transaction) error {
			return nil
		})

		pg8 := monkey.Patch(restRedeemTimes, func(_ *Coupon) (uint, error) {
			return 0, &ErrCouponRulesNoRedeemTimes
		})

		pg9 := monkey.Patch(updateCouponState, func(_ *Coupon) error {
			return &ErrRuleNotFound
		})

		Convey("Should be ErrRuleNotFound", func() {
			_, err := RedeemCouponByType(nil, requester, "abc", "", "")
			So(err, ShouldEqual, &ErrRuleNotFound)
		})
		pg9.Unpatch()
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

	Convey("Given restRedeemTimes great than 0", t, func() {
		var requester = _aRequester(base.RandString(4), []string{}, "some brand")
		pg := monkey.PatchInstanceMethod(reflect.TypeOf(requester), "HasRole", func(_ *base.Requester, _ string) bool {
			return true
		})
		pg2 := monkey.Patch(GetCoupons, func(_ *base.Requester, _ string, _ string) ([]*Coupon, error) {
			c := _aCoupon("abc", "xxx", "yyy", "def", 0, nil)
			return []*Coupon{c}, nil
		})

		pg3 := monkey.Patch(validateCouponRules, func(_ *base.Requester, _ string, _ *Coupon) ([]error, error) {
			return nil, nil
		})

		pg4 := monkey.Patch(validateCouponBasicForRedeem, func(_ *Coupon, _ string) error {
			return nil
		})

		mockSQLTx := new(sql.Tx)
		pg5 := monkey.PatchInstanceMethod(reflect.TypeOf(dbConnection), "Begin", func(*sql.DB) (*sql.Tx, error) {
			return mockSQLTx, nil
		})

		pg6 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Rollback", func(*sql.Tx) error {
			return nil
		})

		pg7 := monkey.Patch(createCouponTransactions, func(_ []*Transaction) error {
			return nil
		})

		pg8 := monkey.Patch(restRedeemTimes, func(_ *Coupon) (uint, error) {
			return 5, nil
		})

		pg9 := monkey.Patch(updateCouponState, func(_ *Coupon) error {
			return &ErrRuleNotFound
		})
		bCommited := false
		pg10 := monkey.PatchInstanceMethod(reflect.TypeOf(mockSQLTx), "Commit", func(*sql.Tx) error {
			bCommited = true
			return nil
		})
		bSendMsg := false
		pg11 := monkey.Patch(_sendRedeemMessage, func([]*Coupon, string) {
			bSendMsg = true
		})

		Convey("Should commited and sent msg", func() {
			_, _ = RedeemCouponByType(nil, requester, "abc", "", "")
			So(bCommited, ShouldBeTrue)
			So(bSendMsg, ShouldBeTrue)
		})

		pg11.Unpatch()
		pg10.Unpatch()
		pg9.Unpatch()
		pg8.Unpatch()
		pg7.Unpatch()
		pg6.Unpatch()
		pg5.Unpatch()
		pg4.Unpatch()
		pg3.Unpatch()
		pg2.Unpatch()
		pg.Unpatch()
	})

}

func Test_validateCouponForRedeem(t *testing.T) {
	Convey("Given a nil coupon", t, func() {
		err := validateCouponBasicForRedeem(nil, "abc")
		So(err, ShouldEqual, &ErrCouponNotFound)
	})

	Convey("Given a SRedeemed coupon", t, func() {
		c := _aCoupon("A", "xxx", "yyy", "B", SRedeemed, nil)
		err := validateCouponBasicForRedeem(c, "abc")
		So(err, ShouldEqual, &ErrCouponWasRedeemed)
	})

	Convey("Given a SRevoked coupon", t, func() {
		c := _aCoupon("A", "xxx", "yyy", "B", SRevoked, nil)
		err := validateCouponBasicForRedeem(c, "abc")
		So(err, ShouldEqual, &ErrCouponIsNotActive)
	})

	Convey("Given a mismatch consumer", t, func() {
		c := _aCoupon("A", "xxx", "yyy", "B", SActive, nil)
		err := validateCouponBasicForRedeem(c, "abc")
		So(err, ShouldEqual, &ErrCouponWrongConsumer)
	})
}
