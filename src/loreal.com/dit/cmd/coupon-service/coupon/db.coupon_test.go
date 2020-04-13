package coupon

import (
	"fmt"
	// "time"

	// "reflect"
	"testing"

	"loreal.com/dit/cmd/coupon-service/base"
	// "loreal.com/dit/utils"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_GetCoupons(t *testing.T) {
	Convey("Given a coupon in an empty table", t, func() {
		c := _prepareARandomCouponInDB(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID, true)
		Convey("Get the coupon and should contain the coupon just created", func() {
			cs, _ := getCoupons(c.ConsumerID, defaultCouponTypeID)
			So(len(cs), ShouldEqual, 1)
			c2 := cs[0]
			So(base.IsValidUUID(c2.ID), ShouldBeTrue)
			So(c.ConsumerID, ShouldEqual, c2.ConsumerID)
			So(c.ConsumerRefID, ShouldEqual, c2.ConsumerRefID)
			So(c.ChannelID, ShouldEqual, c2.ChannelID)
			So(defaultCouponTypeID, ShouldEqual, c2.CouponTypeID)
			So(c.State, ShouldEqual, c2.State)
		})
	})

	// Convey("Given several coupons for 2 consumers in an empty table", t, func() {
	// 	cs := _prepareSeveralCouponsInDB(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID, true)
	// 	_ = _prepareSeveralCouponsInDB(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID, false)
	// 	Convey("Get for the first consumer and should contain all the coupons just created for first consumer", func() {
	// 		cs3, _ := getCoupons(cs[0].ConsumerID, defaultCouponTypeID)
	// 		So(len(cs3), ShouldEqual, len(cs))
	// 		c := cs[r.Intn(len(cs))]
	// 		for _, c2 := range cs3 {
	// 			if c.ID == c2.ID {
	// 				So(c.ConsumerID, ShouldEqual, c2.ConsumerID)
	// 				break
	// 			}
	// 		}
	// 	})
	// })

	// Convey("Given several coupons with different coupon type in an empty table", t, func() {
	// 	consumerID := base.RandString(4)
	// 	cs := _prepareSeveralCouponsInDB(consumerID, base.RandString(4), base.RandString(4), defaultCouponTypeID, true)
	// 	_ = _prepareSeveralCouponsInDB(consumerID, base.RandString(4), base.RandString(4), anotherCouponTypeID, false)
	// 	Convey("Get the coupon with defaultCouponTypeID and only with defaultCouponTypeID", func() {
	// 		cs2, _ := getCoupons(consumerID, defaultCouponTypeID)
	// 		So(len(cs), ShouldEqual, len(cs2))
	// 		c := cs[r.Intn(len(cs))]
	// 		for _, c2 := range cs2 {
	// 			if c.ID == c2.ID {
	// 				So(c.ConsumerID, ShouldEqual, c2.ConsumerID)
	// 				So(c.ConsumerRefID, ShouldEqual, c2.ConsumerRefID)
	// 				So(c.ChannelID, ShouldEqual, c2.ChannelID)
	// 				So(c.CouponTypeID, ShouldEqual, c2.CouponTypeID)
	// 				So(c.State, ShouldEqual, c2.State)
	// 				break
	// 			}
	// 		}
	// 	})
	// })

	// Convey("Given a coupon in an empty table and will query with wrong input", t, func() {
	// 	c := _prepareARandomCouponInDB(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID, true)
	// 	Convey("Get the coupon with wrong type and should contain nothing", func() {
	// 		cs, _ := getCoupons(c.ConsumerID, anotherCouponTypeID)
	// 		So(len(cs), ShouldEqual, 0)
	// 	})

	// 	Convey("Get the coupon with wrong consumerID and should contain nothing", func() {
	// 		wrongConsumerID := "this is not a real type"
	// 		cs, _ := getCoupons(wrongConsumerID, defaultCouponTypeID)
	// 		So(len(cs), ShouldEqual, 0)
	// 	})
	// })
}

func Test_GetCoupon(t *testing.T) {
	Convey("Given a coupon in an empty table", t, func() {
		c := _prepareARandomCouponInDB(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID, true)
		Convey("get this coupon and should be same with the coupon just created", func() {
			c2, _ := getCoupon(c.ID)
			So(c.ID, ShouldEqual, c2.ID)
			So(c.ConsumerID, ShouldEqual, c2.ConsumerID)
			So(c.ConsumerRefID, ShouldEqual, c2.ConsumerRefID)
			So(c.ChannelID, ShouldEqual, c2.ChannelID)
			So(c.CouponTypeID, ShouldEqual, c2.CouponTypeID)
			So(c.State, ShouldEqual, c2.State)
			So(c.Properties, ShouldContainKey, "the_key")
			So(c.Properties["the_key"], ShouldEqual, "the value")
		})
	})

	Convey("Given several coupons in an empty table", t, func() {
		cs := _prepareSeveralCouponsInDB(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID, true)
		Convey("get the random coupon and should be same with the first coupon just created", func() {
			index := r.Intn(len(cs))
			c2, _ := getCoupon(cs[index].ID)
			So(cs[index].ID, ShouldEqual, c2.ID)
			So(cs[index].ConsumerID, ShouldEqual, c2.ConsumerID)
			So(cs[index].ConsumerRefID, ShouldEqual, c2.ConsumerRefID)
			So(cs[index].ChannelID, ShouldEqual, c2.ChannelID)
			So(cs[index].CouponTypeID, ShouldEqual, c2.CouponTypeID)
			So(cs[index].State, ShouldEqual, c2.State)
		})
	})

	Convey("Given a truncated table", t, func() {
		_, _ = dbConnection.Exec("DELETE FROM coupons")
		Convey("get with a non existed coupon id and the coupon should be null", func() {
			c, _ := getCoupon("some_coupon_id")
			So(c, ShouldBeNil)
		})
	})
}

func Test_CreateCoupon(t *testing.T) {
	Convey("Given a coupon", t, func() {
		state := r.Intn(int(SUnknown))
		var p map[string]interface{}
		p = make(map[string]interface{}, 1)
		p["the_key"] = "the value"
		cc := _aCoupon(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID, State(state), p)
		Convey("Save the coupon to database", func() {
			createCoupon(cc)
			Convey("The coupon just created can be queried", func() {
				s := fmt.Sprintf(`SELECT id, consumerID, consumerRefID, channelID, couponTypeID, state, properties, createdTime FROM coupons WHERE id = '%s'`, cc.ID)
				row := dbConnection.QueryRow(s)
				var c Coupon
				var pstr string
				_ = row.Scan(&c.ID, &c.ConsumerID, &c.ConsumerRefID, &c.ChannelID, &c.CouponTypeID, &c.State, &pstr, &c.CreatedTime)
				c.SetPropertiesFromString(pstr)
				c.CreatedTimeToLocal()
				Convey("The coupon queried should be same with original", func() {
					So(cc.ID, ShouldEqual, c.ID)
					So(cc.ConsumerID, ShouldEqual, c.ConsumerID)
					So(cc.ConsumerRefID, ShouldEqual, c.ConsumerRefID)
					So(cc.ChannelID, ShouldEqual, c.ChannelID)
					So(cc.CouponTypeID, ShouldEqual, c.CouponTypeID)
					So(cc.State, ShouldEqual, c.State)
					So(cc.Properties, ShouldContainKey, "the_key")
					So(cc.Properties["the_key"], ShouldEqual, "the value")
				})
			})
		})
	})

	Convey("Given nil coupon", t, func() {
		err := createCoupon(nil)
		Convey("Nothing happened", func() {
			So(err, ShouldBeNil)
		})
	})
}

func Test_CreateCoupons(t *testing.T) {
	Convey("Given several coupons", t, func() {
		cs := _someCoupons(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID)
		Convey("createCoupons save them to database", func() {
			_, _ = dbConnection.Exec("DELETE FROM coupons")
			createCoupons(cs)
			Convey("Saved coupons should same with prepared", func() {
				for _, c := range cs {
					s := fmt.Sprintf(`SELECT id, consumerID, consumerRefID, channelID, couponTypeID, state, properties, createdTime FROM coupons WHERE id = '%s'`, c.ID)
					row := dbConnection.QueryRow(s)
					var c2 Coupon
					var pstr string
					_ = row.Scan(&c2.ID, &c2.ConsumerID, &c2.ConsumerRefID, &c2.ChannelID, &c2.CouponTypeID, &c2.State, &pstr, &c2.CreatedTime)
					c2.SetPropertiesFromString(pstr)
					So(c.ID, ShouldEqual, c2.ID)
					So(c.ConsumerID, ShouldEqual, c2.ConsumerID)
					So(c.ConsumerRefID, ShouldEqual, c2.ConsumerRefID)
					So(c.ChannelID, ShouldEqual, c2.ChannelID)
					So(c.CouponTypeID, ShouldEqual, c2.CouponTypeID)
					So(c.State, ShouldEqual, c2.State)
					So(c2.Properties, ShouldContainKey, "the_key")
					So(c2.Properties["the_key"], ShouldEqual, "the value")
				}
			})
		})
	})

	Convey("Given empty coupon list", t, func() {
		err := createCoupons(make([]*Coupon, 0))
		Convey("Nothing happened", func() {
			So(err, ShouldBeNil)
		})
	})
}

func Test_UpdateCouponState(t *testing.T) {
	Convey("Given a coupon in an empty table", t, func() {
		c := _prepareARandomCouponInDB(base.RandString(4), base.RandString(4), base.RandString(4), defaultCouponTypeID, true)
		Convey("Reset the coupon state", func() {
			oldState := int(c.State)
			c.State = State((oldState + 1) % int(SUnknown))
			newState := int(c.State)
			So(oldState, ShouldNotEqual, int(c.State))
			updateCouponState(c)
			Convey("The latest coupon should have diff state with original", func() {
				c2, _ := getCoupon(c.ID)
				So(c2.State, ShouldEqual, newState)
			})
		})
	})

	Convey("Given nil coupon", t, func() {
		err := updateCouponState(nil)
		Convey("Nothing happened", func() {
			So(err, ShouldBeNil)
		})
	})
}

func Test_CreateCouponTransaction(t *testing.T) {
	Convey("Given a coupon transaction", t, func() {
		tt := r.Intn(int(TTUnknownTransaction) + 1)
		t := _aTransaction(base.RandString(4), uuid.New().String(), TransType(tt), base.RandString(4))
		Convey("Save the transaction to database", func() {
			createCouponTransaction(t)
			Convey("The transaction just created can be queried", func() {
				s := fmt.Sprintf(`SELECT id, couponID, actorID, transType, extraInfo, createdTime FROM couponTransactions WHERE id = '%s'`, t.ID)
				row := dbConnection.QueryRow(s)
				var t2 Transaction
				var ei string
				_ = row.Scan(&t2.ID, &t2.CouponID, &t2.ActorID, &t2.TransType, &ei, &t2.CreatedTime)
				t2.DecryptExtraInfo(ei)
				Convey("The transaction queried should be same with original", func() {
					So(t.ID, ShouldEqual, t2.ID)
					So(t.CouponID, ShouldEqual, t2.CouponID)
					So(t.ActorID, ShouldEqual, t2.ActorID)
					So(t.TransType, ShouldEqual, t2.TransType)
					So(t.ExtraInfo, ShouldEqual, t2.ExtraInfo)
				})
			})
		})
	})

	Convey("Given nil coupon transaction", t, func() {
		err := createCouponTransaction(nil)
		Convey("Nothing happened", func() {
			So(err, ShouldBeNil)
		})
	})
}

func Test_CreateCouponTransactions(t *testing.T) {
	Convey("Given several coupon transactions", t, func() {
		ts := _someTransaction(base.RandString(4), uuid.New().String(), base.RandString(4))
		Convey("createCouponTransactions save them to database", func() {
			_, _ = dbConnection.Exec("DELETE FROM coupons")
			createCouponTransactions(ts)
			Convey("Saved coupon transactions should same with prepared", func() {
				for _, t := range ts {
					s := fmt.Sprintf(`SELECT id, couponID, actorID, transType, extraInfo, createdTime FROM couponTransactions WHERE id = '%s'`, t.ID)
					row := dbConnection.QueryRow(s)
					var t2 Transaction
					var ei string
					_ = row.Scan(&t2.ID, &t2.CouponID, &t2.ActorID, &t2.TransType, &ei, &t2.CreatedTime)
					t2.DecryptExtraInfo(ei)
					So(t.ID, ShouldEqual, t2.ID)
					So(t.CouponID, ShouldEqual, t2.CouponID)
					So(t.ActorID, ShouldEqual, t2.ActorID)
					So(t.TransType, ShouldEqual, t2.TransType)
					So(t.ExtraInfo, ShouldEqual, t2.ExtraInfo)
				}
			})
		})
	})

	Convey("Given empty coupon transaction list", t, func() {
		err := createCouponTransactions(make([]*Transaction, 0))
		Convey("Nothing happened", func() {
			So(err, ShouldBeNil)
		})
	})
}

func Test_GetCouponTransactionsWithType(t *testing.T) {
	Convey("Given a coupon transaction in an empty table", t, func() {
		tt := r.Intn(int(TTUnknownTransaction))
		t := _prepareARandomCouponTransactionInDB(base.RandString(4), uuid.New().String(), TransType(tt), true)
		Convey("getCouponTransactionsWithType can get the coupon transaction", func() {
			ts, _ := getCouponTransactionsWithType(t.CouponID, TransType(tt))
			Convey("And should contain the coupon transaction just created", func() {
				So(len(ts), ShouldEqual, 1)
				t2 := ts[0]
				So(base.IsValidUUID(t2.ID), ShouldBeTrue)
				So(t.CouponID, ShouldEqual, t2.CouponID)
				So(t.ActorID, ShouldEqual, t2.ActorID)
				So(t.TransType, ShouldEqual, t2.TransType)
			})
		})

		Convey("getCouponTransactionsWithType try to get the coupon transaction with diff trans type", func() {
			tt2 := tt + 1%int(TTUnknownTransaction)
			ts, _ := getCouponTransactionsWithType(t.CouponID, TransType(tt2))
			Convey("And should not contain the coupon just created", func() {
				So(len(ts), ShouldEqual, 0)
			})
		})

		Convey("getCouponTransactionsWithType try to get the coupon transaction with wrong coupon id", func() {
			ts, _ := getCouponTransactionsWithType(base.RandString(4), TransType(tt))
			Convey("And should not contain the coupon just created", func() {
				So(len(ts), ShouldEqual, 0)
			})
		})
	})

	Convey("Given several coupon transactions with different coupon in an empty table", t, func() {
		couponID := uuid.New().String()
		tt3 := r.Intn(int(TTUnknownTransaction))
		ts := _prepareSeveralCouponTransactionsInDB(base.RandString(4), couponID, TransType(tt3), true)
		_ = _prepareSeveralCouponTransactionsInDB(base.RandString(4), uuid.New().String(), TransType(tt3), false)
		Convey("getCouponTransactionsWithType only get the coupon transactions with given couponID", func() {
			ts2, _ := getCouponTransactionsWithType(couponID, TransType(tt3))
			Convey("And should contain the coupon ransactions just created with given couponID", func() {
				So(len(ts), ShouldEqual, len(ts2))
				t := ts[0]
				for _, t2 := range ts2 {
					if t.ID == t2.ID {
						So(t.ID, ShouldEqual, t2.ID)
						So(t.CouponID, ShouldEqual, t2.CouponID)
						So(t.ActorID, ShouldEqual, t2.ActorID)
						So(t.TransType, ShouldEqual, t2.TransType)
						break
					}
				}
			})
		})
	})
}

func Test_GetCouponTransactionCountWithType(t *testing.T) {
	Convey("Given a clean db", t, func() {
		dbConnection.Exec("DELETE FROM couponTransactions")
		Convey("There is no coupon transactio", func() {
			tt := r.Intn(int(TTUnknownTransaction))
			count, err := getCouponTransactionCountWithType(base.RandString(4), TransType(tt))
			So(0, ShouldEqual, count)
			So(err, ShouldBeNil)
		})
	})

	Convey("Given several coupon transactions with different trans type in an empty table", t, func() {
		couponID := uuid.New().String()
		tt := r.Intn(int(TTUnknownTransaction))
		ts := _prepareSeveralCouponTransactionsInDB(base.RandString(4), couponID, TransType(tt), true)
		tt2 := (tt + 1) % int(TTUnknownTransaction)
		_ = _prepareSeveralCouponTransactionsInDB(base.RandString(4), couponID, TransType(tt2), false)
		Convey("getCouponTransactionsWithType only get the coupon transactions with given type", func() {
			count, err := getCouponTransactionCountWithType(couponID, TransType(tt))
			Convey("And should contain the coupon ransactions just created with given couponID", func() {
				So(len(ts), ShouldEqual, count)
				So(err, ShouldBeNil)
			})
		})
	})
}

func _prepareSeveralCouponTransactionsInDB(actorID string, couponID string, tt TransType, bTrancate bool) []*Transaction {
	count := r.Intn(10) + 1
	if bTrancate {
		_, _ = dbConnection.Exec("DELETE FROM couponTransactions")
	}
	ts := make([]*Transaction, 0, count)
	for i := 0; i < count; i++ {
		ts = append(ts, _prepareARandomCouponTransactionInDB(actorID, couponID, tt, false))
	}
	return ts
}

func _prepareARandomCouponTransactionInDB(actorID string, couponID string, tt TransType, bTrancate bool) *Transaction {
	t := _aTransaction(actorID, couponID, TransType(tt), base.RandString(4))
	if bTrancate {
		_, _ = dbConnection.Exec("DELETE FROM couponTransactions")
	}
	s := fmt.Sprintf(`INSERT INTO couponTransactions VALUES ("%s","%s","%s",%d,"%s","%d")`, t.ID, t.CouponID, t.ActorID, t.TransType, t.EncryptExtraInfo(), t.CreatedTime.Unix())
	_, e := dbConnection.Exec(s)
	if nil != e {
		fmt.Println("dbConnection.Exec(s) ====  ", e.Error())
	}
	return t
}

func _prepareSeveralCouponsInDB(consumerID string, consumerRefID string, channelID string, couponType string, bTrancate bool) []*Coupon {
	count := r.Intn(10) + 1
	if bTrancate {
		_, _ = dbConnection.Exec("DELETE FROM coupons")
	}
	cs := make([]*Coupon, 0, count)
	for i := 0; i < count; i++ {
		cs = append(cs, _prepareARandomCouponInDB(consumerID, consumerRefID, channelID, couponType, false))
	}
	return cs
}

func _prepareARandomCouponInDB(consumerID string, consumerRefID string, channelID string, couponType string, bTrancate bool) *Coupon {
	state := r.Intn(int(SUnknown))
	var p map[string]interface{}
	p = make(map[string]interface{}, 1)
	p["the_key"] = "the value"
	c := _aCoupon(consumerID, consumerRefID, channelID, couponType, State(state), p)
	if bTrancate {
		_, _ = dbConnection.Exec("DELETE FROM coupons")
	}
	s := fmt.Sprintf(`INSERT INTO coupons VALUES ("%s","%s","%s","%s","%s",%d,"%s",%d)`, c.ID, c.CouponTypeID, c.ConsumerID, c.ConsumerRefID, c.ChannelID, c.State, "some string", c.CreatedTime.Unix())
	_, _ = dbConnection.Exec(s)
	return c
}
