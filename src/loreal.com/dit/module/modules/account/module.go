package account

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" //sqlite3 database driver

	"loreal.com/dit/module"
)

func init() {
	makeDataFolder()
}

func makeDataFolder() {
	path := "./data"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModeDir|0770); err != nil {
				log.Println("[Mkdir data]", err)
			}
		}
	}
}

//Module - Micro service module for root path
type Module struct {
	*module.Module
	LoggedInUsers  *LoggedInUsers
	TokenIssuer    string
	db             *sql.DB
	failLockCount  int
	SignKey        []byte
	unlockIn       time.Duration
	tokenExpiresIn time.Duration
	mutex          *sync.RWMutex
}

//ErrInvalidParameter - invalid parameter
var ErrInvalidParameter = errors.New("Invalid Parameter")

//ErrLoginFailed - incorrect uid or password
var ErrLoginFailed = errors.New("Login Failded")

//ErrAccountLocked - incorrect uid or password
var ErrAccountLocked = errors.New("Account Locked")

//ErrInvalidToken - invalid or expired token
var ErrInvalidToken = errors.New("Invalid Token")

//ErrAccountDisabled - account disabled
var ErrAccountDisabled = errors.New("Account Disabled")

//ErrAccountNotFound - account not found
var ErrAccountNotFound = errors.New("Account Not Found")

//ErrAccountExists - account already exists
var ErrAccountExists = errors.New("Account Exists")

//ErrPasswordRule - Password rule violation
var ErrPasswordRule = errors.New("Violation of password rule")

//ErrPasswordNotMatch - 2 Password don't match
var ErrPasswordNotMatch = errors.New("Password don't match")

//GetBase - implements ISubModule interface
func (m *Module) GetBase() *module.Module {
	return m.Module
}

func fixPrefix(prefix string) string {
	if prefix != "/" {
		return strings.TrimRight("/"+strings.Trim(prefix, "/"), "/")
	}
	return ""
}

//NewModule create new user account module
func NewModule(path string, TokenIssuer string, SignKey []byte, failLockCount int, unlockIn, tokenExpiresIn time.Duration) *Module {
	m := &Module{
		Module: module.NewModule("account", "Hash based user account module", path),
		LoggedInUsers: &LoggedInUsers{
			LoginLimit: DefaultAccountConfig.LoginLimit,
			Logins:     map[string][]time.Time{},
			mutex:      &sync.Mutex{},
		},
		TokenIssuer:    TokenIssuer,
		failLockCount:  failLockCount,
		unlockIn:       unlockIn,
		tokenExpiresIn: tokenExpiresIn,
		mutex:          &sync.RWMutex{},
	}
	m.WatchDogTick = 5 * time.Minute
	m.registerHandlers()
	m.registerEndpoints()

	m.OnInit = func(p *module.Module) {
		m.initDB()
	}
	m.OnTick = func(p *module.Module) {
		//log.Println("\tPurging SMS DB...")
		if _, err := m.db.Exec("delete from accounts where expiresOn <> 0 and expiresOn<?", time.Now()); err != nil {
			log.Print(err)
			return
		}
	}
	m.OnDispose = func(p *module.Module) {
		m.disposeDB()
	}
	return m
}

//disposeDB module resources
func (m *Module) disposeDB() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *Module) initDB() {
	var err error
	m.db, err = sql.Open("sqlite3", "./data/user-accounts.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}

	//init database tables
	sqlStmts := []string{
		"create table if not exists accounts (uid text primary key,hash text,roles text,properties text,publicprops text,failCount int default 0,locked int default 0,enabled int default 1,expiresOn datetime default 0,createAt datetime default 0,modifiedAt datetime default 0,loginAt datetime default 0);",
		"create index if not exists expireOnIdx on accounts(expiresOn);",
	}
	for _, sqlStmt := range sqlStmts {
		_, err := m.db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return
		}
	}
}
