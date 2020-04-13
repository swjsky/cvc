package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

//SignURLValues - sign url values by a key with generated nonce, timestamp
func SignURLValues(args url.Values, key string) (signature, nonce, timestamp string) {
	nonce = RandomString(16)
	timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	signature = sign(args, key, nonce, timestamp)
	return
}

//VerifySignature - Verify signature against input values
func VerifySignature(args url.Values, key, nonce, timestamp, signature string) bool {
	return signature == sign(args, key, nonce, timestamp)
}

//DebugSignature - Verify signature against input values - return signature
func DebugSignature(args url.Values, key, nonce, timestamp, signature string, beforeHash, correctSignature *string) bool {
	*beforeHash, *correctSignature = debugSign(args, key, nonce, timestamp)
	return signature == *correctSignature
}

func debugSign(args url.Values, key, nonce, timestamp string) (beforeHash, signature string) {
	//	log.Println("encode:", args.Encode())
	//	log.Println("str", args.Encode()+key+nonce+timestamp)
	beforeHash = args.Encode() + key + nonce + timestamp
	hashsum := sha1.Sum([]byte(beforeHash))
	signature = hex.EncodeToString(hashsum[:])
	if os.Getenv("EV_DEBUG") != "" {
		log.Println("BeforeHash:", beforeHash, "Signature:", signature)
	}
	return
}

func sign(args url.Values, key, nonce, timestamp string) (signature string) {
	//	log.Println("encode:", args.Encode())
	//	log.Println("str", args.Encode()+key+nonce+timestamp)
	beforeHash := args.Encode() + key + nonce + timestamp
	hashsum := sha1.Sum([]byte(beforeHash))
	signature = hex.EncodeToString(hashsum[:])
	if os.Getenv("EV_DEBUG") != "" {
		log.Println("BeforeHash:", beforeHash, "Signature:", signature)
	}
	return
}
