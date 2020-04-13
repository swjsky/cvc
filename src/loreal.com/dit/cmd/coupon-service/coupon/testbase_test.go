package coupon

import (
	"database/sql"
	"os"

	// "fmt"
	"math/rand"

	// "reflect"
	"testing"
	"time"

	"loreal.com/dit/cmd/coupon-service/base"
	"loreal.com/dit/utils"

	"github.com/google/uuid"
	migrate "github.com/rubenv/sql-migrate"
)

var r *rand.Rand = rand.New(rand.NewSource(time.Now().Unix()))

const defaultCouponTypeID string = "678719f5-44a8-4ac8-afd0-288d2f14daf8"
const anotherCouponTypeID string = "dff0710e-f5af-4ecf-a4b5-cc5599d98030"

func TestMain(m *testing.M) {
	_setUp()
	m.Run()
	_tearDown()
}

func _setUp() {
	encryptKey = []byte("a9ad231b0f2a4f448b8846fd1f57813a")
	err := os.Remove("../data/testdata.sqlite")
	dbConnection, err = sql.Open("sqlite3", "../data/testdata.sqlite?cache=shared&mode=rwc")
	if err != nil {
		panic(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: "../sql-migrations",
	}
	_, err = migrate.Exec(dbConnection, "sqlite3", migrations, migrate.Up)
	if err != nil {
		panic(err)
	}
	dbInit()

	utils.LoadOrCreateJSON("./test/rules.json", &supportedRules)
	ruleInit(supportedRules)
}

func _tearDown() {
	os.Remove("./data/testdata.sqlite")
	dbConnection.Close()
}

func _aCoupon(consumerID string, consumerRefID string, channelID string, couponTypeID string, state State, p map[string]interface{}) *Coupon {
	lt := time.Now().Local()

	var c = Coupon{
		ID:            uuid.New().String(),
		CouponTypeID:  couponTypeID,
		ConsumerID:    consumerID,
		ConsumerRefID: consumerRefID,
		ChannelID:     channelID,
		State:         state,
		Properties:    p,
		CreatedTime:   &lt,
	}
	return &c
}

func _someCoupons(consumerID string, consumerRefID string, channelID string, couponTypeID string) []*Coupon {
	count := r.Intn(10) + 1
	cs := make([]*Coupon, 0, count)
	for i := 0; i < count; i++ {
		state := r.Intn(int(SUnknown))
		var p map[string]interface{}
		p = make(map[string]interface{}, 1)
		p["the_key"] = "the value"
		cs = append(cs, _aCoupon(consumerID, consumerRefID, channelID, couponTypeID, State(state), p))
	}
	return cs
}

func _aTransaction(actorID string, couponID string, tt TransType, extraInfo string) *Transaction {
	var t = Transaction{
		ID:          uuid.New().String(),
		CouponID:    couponID,
		ActorID:     actorID,
		TransType:   tt,
		ExtraInfo:   extraInfo,
		CreatedTime: time.Now().Local(),
	}
	return &t
}

func _someTransaction(actorID string, couponID string, extraInfo string) []*Transaction {
	count := r.Intn(10) + 1
	ts := make([]*Transaction, 0, count)
	for i := 0; i < count; i++ {
		tt := r.Intn(int(TTUnknownTransaction))
		ts = append(ts, _aTransaction(actorID, couponID, TransType(tt), extraInfo))
	}
	return ts
}

func _aRequester(userID string, roles []string, brand string) *base.Requester {
	var requester base.Requester
	requester.UserID = userID
	requester.Roles = map[string]([]string){
		"roles": roles,
	}
	requester.Brand = brand
	return &requester
}
