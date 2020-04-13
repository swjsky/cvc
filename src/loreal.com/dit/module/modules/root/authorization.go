package root

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"time"

	"loreal.com/dit/utils"
)

//Account - user account
type Account struct {
	UID    string `json:"uid"`
	Hash   string `json:"hash"`
	Hashed bool   `json:"hashed"`
}

//Accounts - user accounts for module
var Accounts *[]Account

func encodingMd5(password string) string {
	return fmt.Sprintf("%X", md5.Sum([]byte(password)))
}

func init() {
	makeConfigFolder()
	reloadAccounts()
}

func makeConfigFolder() {
	path := "./config"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModeDir|0770); err != nil {
				log.Println("[Mkdir config]", err)
			}
		}
	}
}

func reloadAccounts() {
	var err error
	const fName = "./config/accounts.json"
	Accounts = new([]Account)
	*Accounts = append(*Accounts, Account{
		UID:    "admin",
		Hash:   encodingMd5("admin"),
		Hashed: true,
	})
	if err = utils.LoadOrCreateJSON(fName, &Accounts); err != nil {
		log.Printf("Error when Load JSON file [%s]: %s\r\n", fName, err)
	}
	//log.Println(Accounts)
	var modified bool
	newAccounts := new([]Account)
	for _, a := range *Accounts {
		if !a.Hashed {
			a.Hash = encodingMd5(a.Hash)
			a.Hashed = true
			modified = true
		}
		*newAccounts = append(*newAccounts, a)
	}
	if modified {
		utils.SaveJSON(newAccounts, fName)
	}
	Accounts = newAccounts
}

//FindToken implements TokenLocater interface
func (m *Module) FindToken(token, realm string, callback func(found bool, expireAt time.Time)) {
	log.Println(token, realm)
	if callback != nil {
		callback(true, time.Now().Add(120*time.Minute))
	}

}

//VerifyPassword implements UserVerifier interface
func (m *Module) VerifyPassword(username, password, realm string) bool {
	for _, a := range *Accounts {
		if a.UID != username {
			continue
		}
		switch {
		case a.Hashed && a.Hash == encodingMd5(password):
			return true
		case !a.Hashed && a.Hash == password:
			return true
		default:
			return false
		}
	}
	return false
}
