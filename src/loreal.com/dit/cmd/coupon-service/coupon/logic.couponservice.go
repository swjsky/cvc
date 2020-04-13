package coupon

import (
	"database/sql"
	"errors"
	"strconv"

	// "fmt"
	"log"
	// "net/http"
	"strings"
	"sync"
	"time"

	"loreal.com/dit/cmd/coupon-service/base"
	"loreal.com/dit/utils"

	"github.com/google/uuid"
)

var bufferSize int = 5
var templates []*Template
var publishedCouponTypes []*PublishedCouponType
var encryptKey []byte

// 做一下访问控制
var couponMessageChannels map[chan Message]bool
var serviceMutex *sync.RWMutex

// Init 初始化一些数据
// TODO: 目前是hard code的数据，后期要重构
func Init(temps []*Template,
	couponTypes []*PublishedCouponType,
	rules []*Rule,
	databaseConnection *sql.DB,
	key string) {
	templates = temps
	publishedCouponTypes = couponTypes
	couponMessageChannels = make(map[chan Message]bool)
	encryptKey = []byte(key)
	serviceMutex = new(sync.RWMutex)
	staticsInit(databaseConnection)
	ruleInit(rules)
	dbInit()
}

// ActivateTestedCoupontypes 激活测试用的卡券
// TODO: 实现卡券类型的相关API后，将会移除
func ActivateTestedCoupontypes() {
	var cts []*PublishedCouponType
	utils.LoadOrCreateJSON("coupon/test/coupon_types.json", &cts)
	for _, t := range cts {
		t.InitRules()
	}
	publishedCouponTypes = cts
	log.Printf("[GetCoupons] 新的publishedCouponTypes: %#v\n", publishedCouponTypes)
}

func _checkCouponType(couponTypeID string) *PublishedCouponType {
	var pct *PublishedCouponType
	for _, value := range publishedCouponTypes {
		if value.ID == couponTypeID {
			pct = value
			return pct
		}
	}
	return nil
}

// GetCouponTypes 获取卡券类型列表
// TODO: 加上各种过滤条件
func GetCouponTypes(requester *base.Requester) ([]*PublishedCouponType, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[GetCoupons] 未知错误: %#v\n", v)
		}
	}()

	if !requester.HasRole(base.ROLE_COUPON_ISSUER) && !requester.HasRole(base.ROLE_COUPON_REDEEMER) {
		return nil, &ErrRequesterForbidden
	}
	// if base.IsBlankString(consumerID) {
	// 	return nil, &ErrConsumerIDInvalid
	// }

	return publishedCouponTypes, nil
}

// IssueCoupons 批量签发卡券
// TODO: 测试下单次调用多少用户量合适
func IssueCoupons(requester *base.Requester, consumerIDs string, consumerRefIDs string, channelID string, couponTypeID string) (*[]*Coupon, map[string][]error, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[IssueCoupons] 未知错误: %#v\n", v)
		}
	}()

	if !requester.HasRole(base.ROLE_COUPON_ISSUER) {
		return nil, nil, &ErrRequesterForbidden
	}

	// 1.分拆 consumers
	consumerIDArray := strings.Split(consumerIDs, ",")
	consumerIDArray = removeDuplicatedConsumers(consumerIDArray)
	var consumerRefIDArray []string = nil
	if !base.IsBlankString(consumerRefIDs) {
		consumerRefIDArray = strings.Split(consumerRefIDs, ",")
		consumerRefIDArray = removeDuplicatedConsumers(consumerRefIDArray)
	}

	if consumerRefIDArray != nil && len(consumerIDArray) != len(consumerRefIDArray) {
		return nil, nil, &ErrConsumerIDsAndRefIDsMismatch
	}

	// 2. 查询coupontype
	// TODO: 未来改成数据库模式，此处要重构
	pct := _checkCouponType(couponTypeID)
	if nil == pct {
		return nil, nil, &ErrCouponTemplateNotFound
	}

	// 3. 校验申请条件
	// TODO: 判断是否有重复的rule

	// 用来记录所有人的规则错误
	// TODO: 此处尝试用 gorouting
	// TODO: 记得测试有很多错误的情况下，前端得到什么。
	var allConsumersRuleCheckErrors map[string][]error = make(map[string][]error, 0)
	for _, consumerID := range consumerIDArray {
		rerrs, rerr := validateTemplateRules(consumerID, couponTypeID, pct)
		if rerr != nil {
			switch rerr {
			case &ErrRuleNotFound:
				{
					log.Println(ErrRuleNotFound)
					return nil, nil, &ErrRuleNotFound
				}
			case &ErrRuleJudgeNotFound:
				{
					log.Println(ErrRuleJudgeNotFound)
					return nil, nil, &ErrRuleJudgeNotFound
				}
			default:
				{
					log.Println("未知校验规则错误")
					return nil, nil, errors.New("内部错误")
				}
			}
		}
		if len(rerrs) > 0 {
			allConsumersRuleCheckErrors[consumerID] = rerrs
		}
	}
	if len(allConsumersRuleCheckErrors) > 0 {
		return nil, allConsumersRuleCheckErrors, nil
	}

	// 4. issue
	lt := time.Now().Local()
	composedRules, err := marshalCouponRules(requester, couponTypeID, pct.Rules)
	if nil != err {
		return nil, nil, err
	}

	var newCs []*Coupon = make([]*Coupon, 0, len(consumerIDArray))
	for idx, consumerID := range consumerIDArray {
		refID := ""
		if consumerRefIDArray != nil {
			refID = consumerRefIDArray[idx]
		}

		var newC = Coupon{
			ID:            uuid.New().String(),
			CouponTypeID:  couponTypeID,
			ConsumerID:    consumerID,
			ConsumerRefID: refID,
			ChannelID:     channelID,
			State:         SActive,
			Properties:    make(map[string]interface{}),
			CreatedTime:   &lt,
			Transactions:  make([]*Transaction, 0),
		}
		newC.Properties[KeyBindingRuleProperties] = composedRules
		newCs = append(newCs, &newC)
	}

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Println(err)
		return nil, nil, errors.New("内部错误")
	}

	err = createCoupons(newCs)
	if nil != err {
		tx.Rollback()
		return nil, nil, err
	}

	// 5. 记录log
	var ts []*Transaction = make([]*Transaction, 0, len(newCs))
	for _, c := range newCs {
		var t = Transaction{
			ID:          uuid.New().String(),
			CouponID:    c.ID,
			ActorID:     requester.UserID,
			TransType:   TTIssueCoupon,
			ExtraInfo:   "",
			CreatedTime: time.Now().Local(),
		}
		ts = append(ts, &t)
	}

	err = createCouponTransactions(ts)
	if nil != err {
		tx.Rollback()
		return nil, nil, err
	}

	tx.Commit()

	return &newCs, nil, nil
}

// IssueCoupon 签发一个卡券
// func IssueCoupon(requester *base.Requester, consumerID string, couponTypeID string) (*Coupon, []error, error) {
// 	if !requester.HasRole(base.ROLE_COUPON_ISSUER) {
// 		return nil, nil, &ErrRequesterForbidden
// 	}

// 	//1. 查询coupontype
// 	// TODO: 未来改成数据库模式，此处要重构
// 	pct := _checkCouponType(couponTypeID)
// 	if nil == pct {
// 		return nil, nil, &ErrCouponTemplateNotFound
// 	}

// 	// 2. 校验申请条件
// 	// TODO: 判断是否有重复的rule
// 	rerrs, rerr := validateTemplateRules(consumerID, couponTypeID, pct)
// 	if rerr != nil {
// 		switch rerr {
// 		case &ErrRuleNotFound:
// 			{
// 				log.Println(ErrRuleNotFound)
// 				return nil, nil, &ErrRuleNotFound
// 			}
// 		case &ErrRuleJudgeNotFound:
// 			{
// 				log.Println(ErrRuleJudgeNotFound)
// 				return nil, nil, &ErrRuleJudgeNotFound
// 			}
// 		default:
// 			{
// 				log.Println("未知校验规则错误")
// 				return nil, nil, errors.New("内部错误")
// 			}
// 		}
// 	}
// 	if len(rerrs) > 0 {
// 		return nil, rerrs, nil
// 	}

// 	// 3. issue
// 	lt := time.Now().Local()
// 	var newC = Coupon{
// 		ID:           uuid.New().String(),
// 		CouponTypeID: couponTypeID,
// 		ConsumerID:   consumerID,
// 		State:        SActive,
// 		Properties:   make(map[string]string),
// 		CreatedTime:  &lt,
// 	}

// 	composedRules, err := marshalCouponRules(requester, couponTypeID, pct.Rules)
// 	if nil != err {
// 		return nil, nil, err
// 	}
// 	newC.Properties[KeyBindingRuleProperties] = composedRules

// 	tx, err := dbConnection.Begin()
// 	if err != nil {
// 		log.Println(err)
// 		return nil, nil, errors.New("内部错误")
// 	}

// 	err = createCoupon(&newC)
// 	if nil != err {
// 		tx.Rollback()
// 		return nil, nil, err
// 	}

// 	// 4. 记录log
// 	var t = Transaction{
// 		ID:          uuid.New().String(),
// 		CouponID:    newC.ID,
// 		ActorID:     requester.UserID,
// 		TransType:   TTIssueCoupon,
// 		CreatedTime: time.Now().Local(),
// 	}
// 	err = createCouponTransaction(&t)
// 	if nil != err {
// 		tx.Rollback()
// 		return nil, nil, err
// 	}
// 	tx.Commit()
// 	return &newC, nil, nil
// }

// ValidateCouponExpired - Valiete whether request coupon is expired.
func ValidateCouponExpired(requester *base.Requester, c *Coupon) (bool, error) {
	return validateCouponExpired(requester, c.ConsumerID, c)
}

// GetCoupon 获取一个卡券的信息
func GetCoupon(requester *base.Requester, id string) (*Coupon, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[GetCoupon] 未知错误: %#v\n", v)
		}
	}()

	if !requester.HasRole(base.ROLE_COUPON_ISSUER) && !requester.HasRole(base.ROLE_COUPON_REDEEMER) {
		return nil, &ErrRequesterForbidden
	}
	if base.IsBlankString(id) {
		return nil, &ErrCouponIDInvalid
	}

	c, err := getCoupon(id)
	if nil != err {
		return nil, err
	}

	ts := make([]*Transaction, 0)
	c.Transactions = ts

	return c, nil
}

// GetCouponWithTransactions - Get coupon by specified couponID and return associated transactions with specified transType in the same time.
func GetCouponWithTransactions(requester *base.Requester, id string, transType string) (*Coupon, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[GetCoupon] 未知错误: %#v\n", v)
		}
	}()

	c, err := GetCoupon(requester, id)
	if nil != err {
		return nil, err
	}

	transTypeID, err := strconv.Atoi(transType)
	if err != nil {
		return nil, err
	}

	ts := make([]*Transaction, 0)
	ts, err = getCouponTransactionsWithType(id, TransType(transTypeID))

	if err != nil {
		return nil, err
	}

	c.Transactions = ts

	return c, nil
}

// GetCoupons 搜索用户所有的卡券
// TODO: 增加各种过滤条件
func GetCoupons(requester *base.Requester, consumerID string, couponTypeID string) ([]*Coupon, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[GetCoupons] 未知错误: %#v\n", v)
		}
	}()

	if !requester.HasRole(base.ROLE_COUPON_ISSUER) && !requester.HasRole(base.ROLE_COUPON_REDEEMER) {
		return nil, &ErrRequesterForbidden
	}
	if base.IsBlankString(consumerID) {
		return nil, &ErrConsumerIDInvalid
	}

	cs, err := getCoupons(consumerID, couponTypeID)
	if nil != err {
		return nil, err
	}

	return cs, nil
}

// GetCouponsWithTransactions - Get conpons and specified type of transactions
func GetCouponsWithTransactions(requester *base.Requester, consumerID string, couponTypeID string, transTypeID string) ([]*Coupon, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[GetCoupons] 未知错误: %#v\n", v)
		}
	}()

	cs, err := GetCoupons(requester, consumerID, couponTypeID)

	if err != nil {
		return nil, err
	}

	for _, c := range cs {
		tt, err := strconv.Atoi(transTypeID)
		if err != nil {
			return nil, err
		}

		ts, err := getCouponTransactionsWithType(c.ID, TransType(tt))
		if err != nil {
			return nil, err
		}
		if len(ts) == 0 {
			c.Transactions = make([]*Transaction, 0)
		}
		c.Transactions = ts
	}

	return cs, nil
}

// RedeemCoupon 兑换卡券
func RedeemCoupon(requester *base.Requester, consumerID string, couponID string, extraInfo string) ([]error, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[RedeemCoupon] 未知错误: %#v\n", v)
		}
	}()

	if !requester.HasRole(base.ROLE_COUPON_REDEEMER) {
		return nil, &ErrRequesterForbidden
	}

	//1. 查询coupon
	c, err := GetCoupon(requester, couponID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = validateCouponBasicForRedeem(c, consumerID)
	if nil != err {
		return nil, err
	}

	// 2. validation rules
	rerrs, rerr := validateCouponRules(requester, consumerID, c)
	if rerr != nil {
		return nil, &ErrRuleNotFound
	}

	if len(rerrs) > 0 {
		return rerrs, nil
	}

	// 3. redeem
	var newT = Transaction{
		ID:          uuid.New().String(),
		CouponID:    couponID,
		ActorID:     requester.UserID,
		TransType:   TTRedeemCoupon,
		ExtraInfo:   extraInfo,
		CreatedTime: time.Now().Local(),
	}

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Println(err)
		return nil, errors.New("内部错误")
	}
	err = createCouponTransaction(&newT)
	if nil != err {
		tx.Rollback()
		return nil, err
	}

	// 4. update coupon state
	var count uint
	count, err = restRedeemTimes(c)
	if nil != err {
		if &ErrCouponRulesNoRedeemTimes == err {
			count = 0
		} else {
			tx.Rollback()
			return nil, err
		}
	}

	if 0 == count {
		// 已经兑换完
		c.State = SRedeemed
		err = updateCouponState(c)
		if nil != err {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	// 5. 发通知
	// TODO: 缓存数据，避免失败后丢失数据，当然也可以采用从数据库捞数据的方式，使用cursor
	allCoupons := make([]*Coupon, 0)
	allCoupons = append(allCoupons, c)
	_sendRedeemMessage(allCoupons, extraInfo)

	return nil, nil
}

func _addErrForConsumer(mps map[string][]error, consumerID string, err error) {
	if mps[consumerID] == nil {
		mps[consumerID] = make([]error, 0, 1)
	}
	mps[consumerID] = append(mps[consumerID], err)
}

// RedeemCouponByType 根据卡券类型兑换卡券，
// 要求消费者针对该类型卡券类型只有一张卡券
func RedeemCouponByType(requester *base.Requester, consumerIDs string, couponTypeID string, extraInfo string) (map[string][]error, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[RedeemCouponByType] 未知错误: %#v\n", v)
		}
	}()

	if !requester.HasRole(base.ROLE_COUPON_REDEEMER) {
		return nil, &ErrRequesterForbidden
	}

	var allConsumersRuleCheckErrors map[string][]error = make(map[string][]error, 0)

	//1. 查询coupon
	var consumerIDArray []string = make([]string, 0)
	if !base.IsBlankString(consumerIDs) {
		consumerIDArray = strings.Split(consumerIDs, ",")
	}

	allCoupons := make([]*Coupon, 0, len(consumerIDArray))
	for _, consumerID := range consumerIDArray {
		cs, err := GetCoupons(requester, consumerID, couponTypeID)
		if err != nil {
			log.Println(err)
			_addErrForConsumer(allConsumersRuleCheckErrors, consumerID, err)
			continue
		}
		if nil == cs || len(cs) == 0 {
			_addErrForConsumer(allConsumersRuleCheckErrors, consumerID, &ErrCouponNotFound)
			continue
		}

		// 目前限制只有一份卡券时才可以兑换
		if len(cs) > 1 {
			_addErrForConsumer(allConsumersRuleCheckErrors, consumerID, &ErrCouponTooMuchToRedeem)
			continue
		}

		c := cs[0]
		allCoupons = append(allCoupons, c)

		err = validateCouponBasicForRedeem(c, consumerID)
		if nil != err {
			_addErrForConsumer(allConsumersRuleCheckErrors, consumerID, err)
			continue
		}

		// 2. validation rules
		rerrs, rerr := validateCouponRules(requester, consumerID, c)
		if rerr != nil {
			_addErrForConsumer(allConsumersRuleCheckErrors, consumerID, &ErrRuleNotFound)
			continue
		}

		if len(rerrs) > 0 {
			allConsumersRuleCheckErrors[consumerID] = append(allConsumersRuleCheckErrors[consumerID], rerrs...)
			continue
		}
	}

	if len(allConsumersRuleCheckErrors) > 0 {
		return allConsumersRuleCheckErrors, nil
	}

	// 3. redeem
	ts := make([]*Transaction, 0, len(allCoupons))
	for _, c := range allCoupons {
		var newT = Transaction{
			ID:          uuid.New().String(),
			CouponID:    c.ID,
			ActorID:     requester.UserID,
			TransType:   TTRedeemCoupon,
			ExtraInfo:   extraInfo,
			CreatedTime: time.Now().Local(),
		}
		ts = append(ts, &newT)
	}

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Println(err)
		return nil, errors.New("内部错误")
	}
	err = createCouponTransactions(ts)
	if nil != err {
		tx.Rollback()
		return nil, err
	}

	// 4. update coupon state
	for _, c := range allCoupons {
		var count uint
		count, err = restRedeemTimes(c)
		if nil != err {
			if &ErrCouponRulesNoRedeemTimes == err {
				count = 0
			} else {
				tx.Rollback()
				return nil, err
			}
		}

		if 0 == count {
			// 已经兑换完
			c.State = SRedeemed
			err = updateCouponState(c)
			if nil != err {
				tx.Rollback()
				return nil, err
			}
		}
	}

	tx.Commit()

	// 5. 发通知
	// TODO: 缓存数据，避免失败后丢失数据，当然也可以采用从数据库捞数据的方式，使用cursor
	_sendRedeemMessage(allCoupons, extraInfo)

	return nil, nil
}

// RedeemCouponsInMaintenance - Redeem coupon in maintenance situation
func RedeemCouponsInMaintenance(requester *base.Requester, couponIDs string, extraInfo string) error {
	if base.IsBlankString(couponIDs) {
		return errors.New("empty couponIDs not allowed")
	}

	couponIDArray := strings.Split(couponIDs, ",")

	var limit int = 50
	var batches int = len(couponIDArray) / limit

	if len(couponIDArray)%limit > 0 {
		batches++
	}

	for i := 0; i < batches; i++ {
		min := i * limit
		max := (i + 1) * limit
		if max > len(couponIDArray) {
			max = len(couponIDArray)
		}
		cslice := couponIDArray[min:max]
		coupons := make([]*Coupon, 0)
		for _, cid := range cslice {
			c, err := GetCoupon(requester, cid)

			if err != nil {
				log.Println(err)
				continue
			}

			coupons = append(coupons, c)
		}
		redeemCoupons(requester, coupons, extraInfo)

	}

	return nil
}

func redeemCoupons(requester *base.Requester, allCoupons []*Coupon, extraInfo string) error {
	ts := make([]*Transaction, 0, len(allCoupons))
	for _, c := range allCoupons {
		var newT = Transaction{
			ID:          uuid.New().String(),
			CouponID:    c.ID,
			ActorID:     requester.UserID,
			TransType:   TTRedeemCoupon,
			ExtraInfo:   extraInfo,
			CreatedTime: time.Now().Local(),
		}
		ts = append(ts, &newT)
	}

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Println(err)
		return errors.New("内部错误")
	}
	err = createCouponTransactions(ts)
	if nil != err {
		tx.Rollback()
		return err
	}

	for _, c := range allCoupons {
		var count uint
		count, err = restRedeemTimes(c)
		if nil != err {
			if &ErrCouponRulesNoRedeemTimes == err {
				count = 0
			} else {
				tx.Rollback()
				return err
			}
		}

		if 0 == count {
			// 已经兑换完
			c.State = SRedeemed
			err = updateCouponState(c)
			if nil != err {
				tx.Rollback()
				return err
			}
		}
	}

	tx.Commit()

	return nil
}

func validateCouponBasicForRedeem(c *Coupon, consumerID string) error {
	if nil == c {
		return &ErrCouponNotFound
	}

	if SRedeemed == c.State {
		return &ErrCouponWasRedeemed
	}

	if SActive != c.State {
		return &ErrCouponIsNotActive
	}

	if c.ConsumerID != consumerID {
		return &ErrCouponWrongConsumer
	}

	return nil
}

//TODO: 关于消息还没有单元测试 []*Coupon
func _sendRedeemMessage(cs []*Coupon, extraInfo string) {
	if nil == cs || len(cs) < 0 {
		return
	}
	m := Message{
		Type: MTRedeemed,
		Payload: RedeemedCoupons{
			ExtraInfo: extraInfo,
			Coupons:   cs,
		},
	}
	serviceMutex.RLock()
	defer serviceMutex.RUnlock()
	for chn := range couponMessageChannels {
		chn <- m
	}
}

// DeleteCoupon 删除一个卡券。
// 注意，这里不是用户自己删除, 自己删除的需要考虑后续的业务场景，比如是否可以重新领取
func DeleteCoupon(id string) (*Coupon, error) {

	return nil, nil
}

// GetLatestCouponMessage 获取最新的卡券信息
// TODO: 可能有多个服务器
// TODO: 使用cursor
func GetLatestCouponMessage(requester *base.Requester) (interface{}, error) {
	defer func() {
		if v := recover(); nil != v {
			log.Printf("[GetLatestCouponMessage] 未知错误: %#v\n", v)
		}
	}()

	if !requester.HasRole(base.ROLE_COUPON_LISTENER) {
		return nil, &ErrRequesterForbidden
	}

	chn := make(chan Message)

	serviceMutex.Lock()
	couponMessageChannels[chn] = true
	serviceMutex.Unlock()

	defer func() {
		close(chn)
		serviceMutex.Lock()
		delete(couponMessageChannels, chn)
		serviceMutex.Unlock()
	}()

	msg := <-chn

	return msg, nil
}

// removeDuplicatedConsumers 批量生成卡券时，移除多余的consumerID和consumerRefID
func removeDuplicatedConsumers(consumers []string) []string {
	result := make([]string, 0, len(consumers))
	temp := map[string]struct{}{}
	for _, item := range consumers {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
