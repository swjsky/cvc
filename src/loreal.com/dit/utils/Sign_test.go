package utils

import (
	"log"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestSignURLValues(t *testing.T) {
	var aeskey = "89622015104709087435617163207900"
	var aeskey1 = "89622015104709087435617163207901"
	args := url.Values{}
	args.Add("nnna", "aaab")
	args.Add("a", "b")
	args.Add("nnna2", "190283790187")
	args.Add("url", "https://www.baidu.com/abcd?a=1&b=#2&c=@3#efgh")
	signature, nonce, timestamp := SignURLValues(args, aeskey)
	log.Printf("args=%s, signature=%s, nonce=%s, timestamp=%s", args.Encode(), signature, nonce, timestamp)
	if !VerifySignature(args, aeskey, nonce, timestamp, signature) {
		t.Error("VerifySignature faild case 1, should return true!")
	}
	if VerifySignature(args, aeskey, nonce+"a", timestamp, signature) {
		t.Error("VerifySignature faild case 2, should return false!")
	}
	if VerifySignature(args, aeskey, nonce, strconv.FormatInt(time.Now().Unix()+10, 10), signature) {
		t.Error("VerifySignature faild case 3, should return false!")
	}
	if VerifySignature(args, aeskey1, nonce, timestamp, signature) {
		t.Error("VerifySignature faild case 4, should return false!")
	}
	args.Add("b", "c")
	if VerifySignature(args, aeskey, nonce, timestamp, signature) {
		t.Error("VerifySignature faild case 5, should return false!")
	}
}
