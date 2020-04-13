package utils

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
)

var aeskey = []byte("dfgfd798df7g98dfg7dh9d7/+98fdgjh")

func decodeAESKey(encodedAESKey string) (aesKey []byte, err error) {
	if len(encodedAESKey) != 43 {
		log.Println(encodedAESKey)
		log.Println(len(encodedAESKey))
		return nil, errors.New("the length of encodedAESKey must be equal to 43")
	}
	return base64.StdEncoding.DecodeString(encodedAESKey + "=")
}

//AESEncrypt - plainText = any string
func AESEncrypt(plainText string, AESKey []byte) string {
	const (
		BlockSize = 16
		BlockMask = BlockSize - 1 // BlockSize 为 2^n 时, 可以用 mask 获取针对 BlockSize 的余数
	)

	plaintext := []byte(plainText)
	contentLen := len(plainText)
	amountToPad := BlockSize - contentLen&BlockMask
	bufferLen := contentLen + amountToPad

	buffer := make([]byte, bufferLen)
	copy(buffer, plaintext)

	// 补位 PaddingMode.Zeros
	for i := contentLen; i < bufferLen; i++ {
		buffer[i] = 0
	}

	// 加密
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		panic(err)
	}
	//mode := cipher.NewCBCEncrypter(block, iv)
	mode := NewECBEncrypter(block)

	mode.CryptBlocks(buffer, buffer)

	return base64.StdEncoding.EncodeToString(buffer)
}

//AESDecrypt - ciphertext = any string
func AESDecrypt(encryptedMessage string, AESKey []byte) (plainText string, err error) {
	const (
		BlockSize = 16
		BlockMask = BlockSize - 1 // BlockSize 为 2^n 时, 可以用 mask 获取针对 BlockSize 的余数
	)

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedMessage)

	if len(ciphertext)%BlockSize != 0 {
		return "", fmt.Errorf("crypto/cipher: input not full blocks")
	}

	plaintext := make([]byte, len(ciphertext)) // len(plaintext) >= BlockSize
	if err != nil {
		log.Println("[ERR]AESDecrypt.DecodeString:", err)
		return
	}

	// 解密
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return "", err
	}

	//mode := cipher.NewCBCDecrypter(block, iv)
	mode := NewECBDecrypter(block)
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除补位
	plaintext = bytes.TrimRight(plaintext, string(0))

	return string(plaintext), nil
}

//AESURLEncrypt - plainText = any string
func AESURLEncrypt(plainText string, AESKey []byte) string {
	const (
		BlockSize = 16
		BlockMask = BlockSize - 1 // BlockSize 为 2^n 时, 可以用 mask 获取针对 BlockSize 的余数
	)

	plaintext := []byte(plainText)
	contentLen := len(plainText)
	amountToPad := BlockSize - contentLen&BlockMask
	bufferLen := contentLen + amountToPad

	buffer := make([]byte, bufferLen)
	copy(buffer, plaintext)

	// 补位 PaddingMode.Zeros
	for i := contentLen; i < bufferLen; i++ {
		buffer[i] = 0
	}

	// 加密
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		panic(err)
	}
	//mode := cipher.NewCBCEncrypter(block, iv)
	mode := NewECBEncrypter(block)

	mode.CryptBlocks(buffer, buffer)

	return base64.URLEncoding.EncodeToString(buffer)
}

//AESURLDecrypt - ciphertext = any string
func AESURLDecrypt(encryptedMessage string, AESKey []byte) (plainText string, err error) {
	const (
		BlockSize = 16
		BlockMask = BlockSize - 1 // BlockSize 为 2^n 时, 可以用 mask 获取针对 BlockSize 的余数
	)

	ciphertext, err := base64.URLEncoding.DecodeString(encryptedMessage)

	if len(ciphertext)%BlockSize != 0 {
		return "", fmt.Errorf("crypto/cipher: input not full blocks")
	}

	plaintext := make([]byte, len(ciphertext)) // len(plaintext) >= BlockSize
	if err != nil {
		log.Println("[ERR]AESDecrypt.DecodeString:", err)
		return
	}

	// 解密
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return "", err
	}

	//mode := cipher.NewCBCDecrypter(block, iv)
	mode := NewECBDecrypter(block)
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除补位
	plaintext = bytes.TrimRight(plaintext, string(0))

	return string(plaintext), nil
}

//AES256URLEncrypt - plainText = any string
func AES256URLEncrypt(plainText string, AESKey []byte) string {
	const (
		BlockSize = 32
		BlockMask = BlockSize - 1 // BlockSize 为 2^n 时, 可以用 mask 获取针对 BlockSize 的余数
	)

	plaintext := []byte(plainText)
	contentLen := len(plainText)
	amountToPad := BlockSize - contentLen&BlockMask
	bufferLen := contentLen + amountToPad

	buffer := make([]byte, bufferLen)
	copy(buffer, plaintext)

	// 补位 PaddingMode.Zeros
	for i := contentLen; i < bufferLen; i++ {
		buffer[i] = 0
	}

	// 加密
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		panic(err)
	}
	//mode := cipher.NewCBCEncrypter(block, iv)
	mode := NewECBEncrypter(block)

	mode.CryptBlocks(buffer, buffer)

	return base64.URLEncoding.EncodeToString(buffer)
}

//AES256URLDecrypt - ciphertext = any string
func AES256URLDecrypt(encryptedMessage string, AESKey []byte) (plainText string, err error) {
	const (
		BlockSize = 32
		BlockMask = BlockSize - 1 // BlockSize 为 2^n 时, 可以用 mask 获取针对 BlockSize 的余数
	)

	ciphertext, err := base64.URLEncoding.DecodeString(encryptedMessage)

	if len(ciphertext)%BlockSize != 0 {
		return "", fmt.Errorf("crypto/cipher: input not full blocks")
	}

	plaintext := make([]byte, len(ciphertext)) // len(plaintext) >= BlockSize
	if err != nil {
		log.Println("[ERR]AESDecrypt.DecodeString:", err)
		return
	}

	// 解密
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return "", err
	}

	//mode := cipher.NewCBCDecrypter(block, iv)
	mode := NewECBDecrypter(block)
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除补位
	plaintext = bytes.TrimRight(plaintext, string(0))

	return string(plaintext), nil
}

//AES256Encrypt - plainText = any string
func AES256Encrypt(plainText []byte, AESKey []byte) []byte {
	const (
		BlockSize = 32
		BlockMask = BlockSize - 1 // BlockSize 为 2^n 时, 可以用 mask 获取针对 BlockSize 的余数
	)

	contentLen := len(plainText)
	amountToPad := BlockSize - contentLen&BlockMask
	bufferLen := contentLen + amountToPad

	buffer := make([]byte, bufferLen)
	copy(buffer, plainText)

	// 补位 PaddingMode.Zeros
	for i := contentLen; i < bufferLen; i++ {
		buffer[i] = 0
	}

	// 加密
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		panic(err)
	}
	//mode := cipher.NewCBCEncrypter(block, iv)
	mode := NewECBEncrypter(block)

	mode.CryptBlocks(buffer, buffer)

	return buffer
}

//AES256Decrypt - ciphertext = any string
func AES256Decrypt(ciphertext []byte, AESKey []byte) (plainText []byte, err error) {
	const (
		BlockSize = 32
		BlockMask = BlockSize - 1 // BlockSize 为 2^n 时, 可以用 mask 获取针对 BlockSize 的余数
	)

	if len(ciphertext)%BlockSize != 0 {
		return nil, fmt.Errorf("crypto/cipher: input not full blocks")
	}

	plaintext := make([]byte, len(ciphertext)) // len(plaintext) >= BlockSize
	if err != nil {
		log.Println("[ERR]AESDecrypt.DecodeString:", err)
		return
	}

	// 解密
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return nil, err
	}

	//mode := cipher.NewCBCDecrypter(block, iv)
	mode := NewECBDecrypter(block)
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除补位
	plaintext = bytes.TrimRight(plaintext, string(0))

	return plaintext, nil
}
