package utils

import (
	"log"
	"testing"
)

func TestAESEncrypt(t *testing.T) {

	//var iv = []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF}
	var aeskey = []byte("89622015104709087435617163207900")

	s1 := AESEncrypt("123456", aeskey)
	s2, _ := AESDecrypt(s1, aeskey)
	log.Println(s1)
	log.Println(s2)
	correct := "0cfAwa9X9XRpr53SKjfiug=="
	if s1 != correct {
		t.Error("AESEncrypt failed", s1, " <> ", correct)
	}
	if s2 != "123456" {
		t.Error("AESDecrypt failed", s2, " <> ", "123456")
	}

}

func TestAESEncrypt1(t *testing.T) {

	//var iv = []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF}
	var aeskey = []byte("dfgfd798df7g98dfg7dh9d7/+98fdgjh")

	s1 := AESEncrypt("123456", aeskey)
	s2, _ := AESDecrypt(s1, aeskey)
	// correct := "0cfAwa9X9XRpr53SKjfiug=="
	// if s1 != correct {
	// 	t.Error("AESEncrypt failed", s1, " <> ", correct)
	// }
	log.Println(s1)
	log.Println(s2)
	if s2 != "123456" {
		t.Error("AESDecrypt failed", s2, " <> ", "123456")
	}

}

func TestAES256Encrypt(t *testing.T) {

	//var iv = []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF}
	var aeskey = []byte("dd4be167975e4744a97ffa927daf33d7")

	s1 := AES256URLEncrypt("123456", aeskey)
	s2, _ := AES256URLDecrypt(s1, aeskey)
	// correct := "0cfAwa9X9XRpr53SKjfiug=="
	// if s1 != correct {
	// 	t.Error("AESEncrypt failed", s1, " <> ", correct)
	// }
	if s2 != "123456" {
		t.Error("AESDecrypt failed", s2, " <> ", "123456")
	}

}

func TestAES256(t *testing.T) {

	//var iv = []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF}
	var aeskey = []byte("dd4be167975e4744a97ffa927daf33d7")

	s1 := AES256Encrypt([]byte("123456"), aeskey)
	s2, _ := AES256Decrypt(s1, aeskey)
	// correct := "0cfAwa9X9XRpr53SKjfiug=="
	// if s1 != correct {
	// 	t.Error("AESEncrypt failed", s1, " <> ", correct)
	// }
	if string(s2) != "123456" {
		t.Error("AESDecrypt failed", s2, " <> ", "123456")
	}

}
