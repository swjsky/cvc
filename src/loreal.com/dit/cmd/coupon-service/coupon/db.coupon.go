package coupon

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	// "time"

	"loreal.com/dit/cmd/coupon-service/base"
	// "loreal.com/dit/cmd/coupon-service/rule"
)

var stmtQueryWithType *sql.Stmt
var stmtQueryCoupon *sql.Stmt
var stmtQueryCoupons *sql.Stmt
var stmtInsertCoupon *sql.Stmt
var stmtUpdateCouponState *sql.Stmt
var stmtInsertCouponTransaction *sql.Stmt
var stmtQueryCouponTransactionsWithType *sql.Stmt
var stmtQueryCouponTransactionCountWithType *sql.Stmt
var dbMutex *sync.RWMutex

func dbInit() {
	var err error
	dbMutex = new(sync.RWMutex)
	stmtQueryWithType, err = dbConnection.Prepare("SELECT id, couponTypeID, consumerRefID, channelID, state, properties, createdTime FROM coupons WHERE consumerID = (?) AND couponTypeID = (?)")
	if err != nil {
		log.Println(err)
		panic(err)
	}

	stmtQueryCoupon, err = dbConnection.Prepare("SELECT id, consumerID, consumerRefID, channelID, couponTypeID, state, properties, createdTime FROM coupons WHERE id = (?)")
	if err != nil {
		log.Println(err)
		panic(err)
	}

	stmtQueryCoupons, err = dbConnection.Prepare("SELECT id, couponTypeID, consumerRefID, channelID, state, properties, createdTime FROM coupons WHERE consumerID = (?)")
	if err != nil {
		log.Println(err)
		panic(err)
	}

	stmtInsertCoupon, err = dbConnection.Prepare("INSERT INTO coupons VALUES (?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Println(err)
		panic(err)
	}

	stmtUpdateCouponState, err = dbConnection.Prepare("UPDATE coupons SET state = (?) WHERE id = (?)")
	if err != nil {
		log.Println(err)
		panic(err)
	}

	stmtInsertCouponTransaction, err = dbConnection.Prepare("INSERT INTO couponTransactions VALUES (?,?,?,?,?,?)")
	if err != nil {
		log.Println(err)
		panic(err)
	}

	stmtQueryCouponTransactionsWithType, err = dbConnection.Prepare("SELECT id, actorID, transType, extraInfo, createdTime FROM couponTransactions WHERE couponID = (?) and transType = (?)")
	if err != nil {
		log.Println(err)
		panic(err)
	}

	stmtQueryCouponTransactionCountWithType, err = dbConnection.Prepare("SELECT count(1) FROM couponTransactions WHERE couponID = (?) and transType = (?)")
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

// getCoupons
// TODO: 增加各种过滤条件
func getCoupons(consumerID string, couponTypeID string) ([]*Coupon, error) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	var rows *sql.Rows
	var err error
	if base.IsBlankString(couponTypeID) {
		rows, err = stmtQueryCoupons.Query(consumerID)
	} else {
		rows, err = stmtQueryWithType.Query(consumerID, couponTypeID)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	return getCouponsFromRows(consumerID, rows)
}

func getCouponsFromRows(consumerID string, rows *sql.Rows) ([]*Coupon, error) {
	var coupons []*Coupon = make([]*Coupon, 0)
	for rows.Next() {
		var c Coupon
		var pstr string
		err := rows.Scan(&c.ID, &c.CouponTypeID, &c.ConsumerRefID, &c.ChannelID, &c.State, &pstr, &c.CreatedTime)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		c.SetPropertiesFromString(pstr)
		c.CreatedTimeToLocal()
		c.ConsumerID = consumerID
		c.Transactions = make([]*Transaction, 0)
		coupons = append(coupons, &c)
	}
	err := rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return coupons, nil
}

// getCoupon 获取某一卡券
func getCoupon(couponID string) (*Coupon, error) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	row := stmtQueryCoupon.QueryRow(couponID)

	var c Coupon
	var pstr string
	err := row.Scan(&c.ID, &c.ConsumerID, &c.ConsumerRefID, &c.ChannelID, &c.CouponTypeID, &c.State, &pstr, &c.CreatedTime)
	if err != nil {
		if sql.ErrNoRows == err {
			return nil, nil
		}
		log.Println(err)
		return nil, err
	}

	c.SetPropertiesFromString(pstr)
	c.CreatedTimeToLocal()
	return &c, nil
}

// CreateCoupon 保存卡券到数据库
func createCoupon(c *Coupon) error {
	if nil == c {
		return nil
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()
	pstr, err := c.GetPropertiesString()
	if nil != err {
		return err
	}
	res, err := stmtInsertCoupon.Exec(c.ID, c.CouponTypeID, c.ConsumerID, c.ConsumerRefID, c.ChannelID, c.State, pstr, c.CreatedTime.Unix())
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println(res)
	return nil
}

// CreateCoupon 保存卡券到数据库
func createCoupons(cs []*Coupon) error {
	if len(cs) == 0 {
		return nil
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()

	valueStrings := make([]string, 0, len(cs))
	valueArgs := make([]interface{}, 0, len(cs)*8)
	for _, c := range cs {
		valueStrings = append(valueStrings, "(?,?,?,?,?,?,?,?)")
		valueArgs = append(valueArgs, c.ID)
		valueArgs = append(valueArgs, c.CouponTypeID)
		valueArgs = append(valueArgs, c.ConsumerID)
		valueArgs = append(valueArgs, c.ConsumerRefID)
		valueArgs = append(valueArgs, c.ChannelID)
		valueArgs = append(valueArgs, c.State)
		pstr, err := c.GetPropertiesString()
		if nil != err {
			return err
		}
		valueArgs = append(valueArgs, pstr)
		valueArgs = append(valueArgs, c.CreatedTime.Unix())
	}
	stmt := fmt.Sprintf("INSERT INTO coupons VALUES %s", strings.Join(valueStrings, ","))
	res, err := dbConnection.Exec(stmt, valueArgs...)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(res)
	return nil
}

// updateCouponState 更新卡券状态
func updateCouponState(c *Coupon) error {
	if nil == c {
		return nil
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()

	res, err := stmtUpdateCouponState.Exec(c.State, c.ID)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(res)
	return nil
}

// CreateCouponTransaction 保存卡券操作log到数据库
func createCouponTransaction(t *Transaction) error {
	if nil == t {
		return nil
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()
	res, err := stmtInsertCouponTransaction.Exec(t.ID, t.CouponID, t.ActorID, t.TransType, t.EncryptExtraInfo(), t.CreatedTime.Unix())
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(res)
	return nil
}

// createCouponTransactions 保存卡券操作log到数据库
func createCouponTransactions(ts []*Transaction) error {
	if len(ts) == 0 {
		return nil
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()

	valueStrings := make([]string, 0, len(ts))
	valueArgs := make([]interface{}, 0, len(ts)*6)
	for _, t := range ts {
		valueStrings = append(valueStrings, "(?,?,?,?,?,?)")
		valueArgs = append(valueArgs, t.ID)
		valueArgs = append(valueArgs, t.CouponID)
		valueArgs = append(valueArgs, t.ActorID)
		valueArgs = append(valueArgs, t.TransType)
		valueArgs = append(valueArgs, t.EncryptExtraInfo())
		valueArgs = append(valueArgs, t.CreatedTime.Unix())
	}
	stmt := fmt.Sprintf("INSERT INTO couponTransactions VALUES %s", strings.Join(valueStrings, ","))
	res, err := dbConnection.Exec(stmt, valueArgs...)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(res)
	return nil
}

// getCouponTransactionsWithType 获取某一卡券某类型业务的记录
func getCouponTransactionsWithType(couponID string, ttype TransType) ([]*Transaction, error) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	rows, err := stmtQueryCouponTransactionsWithType.Query(couponID, ttype)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var ts []*Transaction = make([]*Transaction, 0)
	for rows.Next() {
		var t Transaction
		var ei string
		err := rows.Scan(&t.ID, &t.ActorID, &t.TransType, &ei, &t.CreatedTime)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		t.CreatedTimeToLocal()
		err = t.DecryptExtraInfo(ei)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		t.CouponID = couponID
		ts = append(ts, &t)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return ts, nil
}

// getCouponTransactionCountWithType 获取某一卡券某一类型业务的次数，比如查询兑换多少次这样的场景
func getCouponTransactionCountWithType(couponID string, ttype TransType) (uint, error) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	var count uint
	err := stmtQueryCouponTransactionCountWithType.QueryRow(couponID, ttype).Scan(&count)
	if err != nil {
		if sql.ErrNoRows == err {
			return 0, nil
		}
		log.Println(err)
		return 0, err
	}
	return count, nil
}
