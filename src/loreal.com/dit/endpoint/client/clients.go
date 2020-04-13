package client

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"loreal.com/dit/utils"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

//Accounts - Array of ClientAccount
type Accounts []*Account

//DEBUG - whether in debug mode
var DEBUG bool

//INFOLEVEL - info level for debug mode
var INFOLEVEL int

//LOGLEVEL - info level for logs
var LOGLEVEL int

func init() {
	if os.Getenv("EV_DEBUG") != "" {
		DEBUG = true
	}
	INFOLEVEL = 1
	LOGLEVEL = 1
}

//Account - client parameters
type Account struct {
	APPID string `json:"appid"`
	Key   string `json:"key"`
}

//SignURL - sign parameters and generate URL
func (c *Account) SignURL(target *url.URL) {
	params := target.Query()
	nonce, timestamp, signature := c.Sign(params)
	params.Add("appid", c.APPID)
	params.Add("nonce", nonce)
	params.Add("timestamp", timestamp)
	params.Add("signature", signature)
	target.RawQuery = params.Encode()
}

//VerifySignature - verify signature for url parameters
func VerifySignature(clients Accounts, target *url.URL) bool {
	if target == nil {
		return false
	}
	params := target.Query()
	c := clients.Get(params.Get("appid"))
	if c == nil {
		return false
	}
	signature := params.Get("signature")
	if signature == "" {
		return false
	}
	nonce := params.Get("nonce")
	timestamp := params.Get("timestamp")
	params.Del("appid")
	params.Del("signature")
	params.Del("nonce")
	params.Del("timestamp")

	correctSignature := c.sign(nonce, timestamp, params)
	return correctSignature == signature
}

//sign - sign parameters
func (c *Account) sign(nonce, timestamp string, values url.Values) (signature string) {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(values.Encode())
	buffer.WriteString(c.Key)
	buffer.WriteString(nonce)
	buffer.WriteString(timestamp)
	if DEBUG && INFOLEVEL >= 3 {
		log.Println("[DEBUG] - before hash:", buffer.String())
	}
	hash := sha1.Sum(buffer.Bytes())
	signature = hex.EncodeToString(hash[:])
	if DEBUG {
		log.Println("[DEBUG] - signature:", signature)
	}
	return
}

//Sign - sign parameters
func (c *Account) Sign(values url.Values) (nonce, timestamp, signature string) {
	nonce = utils.RandomString(10)
	timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	signature = c.sign(nonce, timestamp, values)
	return
}

//Get - get account by APPID
func (a Accounts) Get(APPID string) *Account {
	for _, account := range a {
		if account.APPID == APPID {
			return account
		}
	}
	return nil
}

//LoadAccounts - load and initialize default accounts
func LoadAccounts(defaultAccounts []*Account) Accounts {
	if defaultAccounts == nil {
		defaultAccounts = []*Account{
			{
				APPID: "sample",
				Key:   "1ca9c7c788a65973ac3f8f0a93c46452",
			},
		}
	}
	clients := Accounts(defaultAccounts)
	utils.MakeFolder("./config/")
	utils.LoadOrCreateJSON("./config/client-accounts.json", &clients)
	return clients
}
