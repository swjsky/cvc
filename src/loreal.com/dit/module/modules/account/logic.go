package account

import (
	"database/sql"
	"log"
	"time"

	"loreal.com/dit/utils"

	"github.com/dgrijalva/jwt-go"
)

const anonymousToken = "053d61e319554343811cba97b8e90352"

//ClaimsWithRoles - Customized Claims
type ClaimsWithRoles struct {
	Roles       string `json:"roles"`
	Props       []byte `json:"props"`
	PublicProps []byte `json:"publicprops"`
	jwt.StandardClaims
}

func (m *Module) newToken(u *Account) (token string) {
	switch {
	case u == nil:
		return
	case u.UID == "anonymous":
		return anonymousToken
	}
	claims := ClaimsWithRoles{
		Roles:       u.Roles,
		Props:       utils.AES256Encrypt(u.PropertiesData, []byte(DefaultAccountConfig.WebTokenKey)),
		PublicProps: u.PublicPropsData,
		StandardClaims: jwt.StandardClaims{
			Id:        u.UID,
			ExpiresAt: time.Now().Add(m.tokenExpiresIn).Unix(), //int64(m.tokenExpiresIn.Seconds()),
			Issuer:    m.TokenIssuer,
		},
	}
	jwToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	token, err := jwToken.SignedString(m.SignKey)
	if err != nil {
		log.Printf("[newToken] error: %v \r\n", err)
	}
	return token
	// conn := m.redisPool.Get()
	// defer conn.Close()
	// token = utils.RandomString(40)
	// _, err := conn.Do("SETEX", token, m.tokenExpiresIn.Seconds(), userAccount.UID+"|"+userAccount.Roles)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
}

//VerifyToken - Verify token, return user account
func (m *Module) VerifyToken(token string) *Account {
	if token == anonymousToken {
		return &Account{
			UID:     "anonymous",
			Roles:   "anonymous",
			Enabled: 1,
			Locked:  0,
		}
	}

	jwToken, err := jwt.ParseWithClaims(token, &ClaimsWithRoles{}, func(token *jwt.Token) (interface{}, error) {
		return m.SignKey, nil
	})
	if err != nil {
		log.Printf("[VerifyToken] Parse jwToken, err: %v", err)
		return nil
	}
	if !jwToken.Valid {
		log.Printf("[VerifyToken] Invalid Token, err: %v", err)
		return nil
	}

	if claims, ok := jwToken.Claims.(*ClaimsWithRoles); ok {
		propertiesData, err := utils.AES256Decrypt(claims.Props, []byte(DefaultAccountConfig.WebTokenKey))
		if err != nil {
			log.Printf("[VerifyToken] Properties decryption err: %v", err)
		}
		u := &Account{
			UID:             claims.Id,
			Roles:           claims.Roles,
			PropertiesData:  propertiesData,
			PublicPropsData: claims.PublicProps,
		}
		u.Parse()
		//log.Printf("[DEBUG] - [VerifyToken][account/logic.go]: %s", string(utils.MarshalJSONV2(u, "  ", false)))
		return u
	}
	log.Printf("[VerifyToken] err: %v", err)
	return nil
}

//Logout - logout by token
// func (m *Module) Logout(token string) (success bool) {
// 	// conn := m.redisPool.Get()
// 	// defer conn.Close()
// 	// _, err := conn.Do("DEL", token)
// 	// if err != nil {
// 	// 	log.Println(err)
// 	// 	return
// 	// }
// 	success = true
// 	return
// }

//Authenticate - by uid and password
func (m *Module) Authenticate(uid, password string) (userAccount *Account, loginErr error) {
	if uid == "anonymous" {
		return &Account{
			UID:     "anonymous",
			Roles:   "anonymous",
			Enabled: 1,
			Locked:  0,
		}, nil
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	userAccount = &Account{
		UID: uid,
	}
	var failCount, enabled int
	var locked int64
	queryErr := m.db.QueryRow("select roles,hash,properties,publicprops,failCount,locked,enabled from accounts where uid=?", uid).Scan(
		&userAccount.Roles,
		&userAccount.HashedPassword,
		&userAccount.PropertiesData,
		&userAccount.PublicPropsData,
		&failCount,
		&locked,
		&enabled)
	switch {
	case queryErr != nil:
		if queryErr == sql.ErrNoRows {
			userAccount = nil
			loginErr = ErrAccountNotFound
			return
		}
		log.Printf("[sqlite3][login][query] error: %v \r\n", queryErr)
		userAccount = nil
		loginErr = ErrAccountNotFound
		return
	case enabled == 0:
		userAccount = nil
		loginErr = ErrAccountDisabled
		return
	case (locked < time.Now().Unix()) && userAccount.VerifyPassword(password): /*Login successfuly*/
		//log.Println("Auth OK, uid =", uid)
		if _, err := m.db.Exec("UPDATE accounts set failCount=0,locked=0,loginAt=? where uid=?;", time.Now(), uid); err != nil {
			log.Printf("[sqlite3][login-success-update] error: %v \r\n", err)
			userAccount = nil
			loginErr = err
			return
		}
		loginErr = userAccount.Parse()
		return
	case failCount+1 > m.failLockCount: /*Login failed - lock*/
		durationToLock := m.unlockIn * time.Duration(failCount+1-m.failLockCount)
		log.Printf("Account Locked for %v, uid = %s", durationToLock, uid)
		_, err := m.db.Exec("UPDATE accounts set locked=?,failCount=failCount+1 where uid=?;",
			time.Now().Add(durationToLock).Unix(),
			uid)
		if err != nil {
			log.Printf("[sqlite3][login-locked-update] error: %v \r\n", err)
			userAccount = nil
			loginErr = err
			return
		}
		loginErr = ErrAccountLocked
		userAccount = nil
		return
	default: /*Login failed - don't need lock*/
		log.Println("Auth. Failed, uid =", uid)
		if _, err := m.db.Exec("UPDATE accounts set failCount=failCount+1 where uid=?;", uid); err != nil {
			log.Printf("[sqlite3][login-failed-update] error: %v \r\n", err)
			userAccount = nil
			loginErr = err
			return
		}
		loginErr = ErrLoginFailed
		userAccount = nil
		return
	}
}

func (m *Module) init(password string) (success bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var rowCount int
	queryErr := m.db.QueryRow("select count(1) from accounts").Scan(&rowCount)
	switch {
	case queryErr != nil:
		log.Printf("[sqlite3][init] error: %v \r\n", queryErr)
		success = false
		return
	case rowCount > 0:
		success = false
		return
	default:
		acc := &Account{
			UID:   "admin",
			Roles: "admin",
		}
		if password == "" {
			password = "admin"
		}
		acc.setPassword(password)
		if _, err := m.db.Exec("INSERT OR REPLACE INTO accounts(uid,hash,roles,createAt) VALUES(?,?,?,?);", acc.UID, acc.HashedPassword, acc.Roles, time.Now()); err != nil {
			log.Printf("[sqlite3][init-upsert] error: %v \r\n", err)
			success = false
			return
		}
		success = true
		return
	}
}

//GetAccount -
func (m *Module) GetAccount(uid string) *Account {
	accounts := m.GetAccounts(uid)
	if len(accounts) == 0 {
		return nil
	}
	return accounts[0]
}

//GetAccounts 获取账户
func (m *Module) GetAccounts(uid string) (accounts []*Account) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var rows *sql.Rows
	var queryErr error

	if uid == "" {
		rows, queryErr = m.db.Query("select uid,roles,properties,publicprops,locked,enabled,createAt,modifiedAt,loginAt,expiresOn from accounts;")
	} else {
		rows, queryErr = m.db.Query("select uid,roles,properties,publicprops,locked,enabled,createAt,modifiedAt,loginAt,expiresOn from accounts where uid=?;", uid)
	}
	if queryErr != nil {
		log.Printf("[sqlite3][get-accounts] error: %v \r\n", queryErr)
		return
	}
	defer rows.Close()
	accounts = make([]*Account, 0)
	for rows.Next() {
		ac := &Account{}
		rowError := rows.Scan(
			&ac.UID,
			&ac.Roles,
			&ac.PropertiesData,
			&ac.PublicPropsData,
			&ac.Locked,
			&ac.Enabled,
			&ac.CreateAt,
			&ac.ModifiedAt,
			&ac.LoginAt,
			&ac.ExpiresOn,
		)
		ac.Parse()
		if rowError == nil {
			accounts = append(accounts, ac)
		} else {
			log.Println("Error:", rowError)
		}
	}
	return
}

//GetPublicProps 获取公有属性
func (m *Module) getPublicProps(uid string) *Account {
	if uid == "" {
		return nil
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	row := m.db.QueryRow("select uid,publicprops from accounts where uid=?;", uid)
	ac := &Account{}
	rowError := row.Scan(
		&ac.UID,
		&ac.PublicPropsData,
	)
	if rowError != nil {
		if rowError != sql.ErrNoRows {
			log.Printf("[sqlite3][getPublicProps] error: %v \r\n", rowError)
		}
		return nil
	}
	ac.ParsePublicProps()
	return ac
}

func (m *Module) insertAccount(userAccount *Account) (insertAccountErr error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var rowCount int
	queryErr := m.db.QueryRow("select count(1) from accounts where uid=?", userAccount.UID).Scan(&rowCount)
	switch {
	case queryErr != nil:
		log.Printf("[sqlite3][save-account-select] error: %v \r\n", queryErr)
		insertAccountErr = queryErr
		return
	case rowCount > 0:
		insertAccountErr = ErrAccountExists
		return
	default:
		if _, err := m.db.Exec("INSERT OR REPLACE INTO accounts(uid,hash,roles,properties,publicprops,createAt) VALUES(?,?,?,?,?,?);",
			userAccount.UID,
			userAccount.HashedPassword,
			userAccount.Roles,
			userAccount.PropertiesData,
			userAccount.PublicPropsData,
			time.Now()); err != nil {
			log.Printf("[sqlite3][save-account-upsert] error: %v \r\n", err)
			insertAccountErr = err
			return
		}
		insertAccountErr = nil
		return
	}
}

func (m *Module) removeAccount(uid string) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if _, err := m.db.Exec("delete from accounts where uid=?", uid); err != nil {
		log.Printf("[sqlite3][removeAccount][delete] error: %v \r\n", err)
		return err
	}
	return nil
}

func (m *Module) updateProperties(userAccount *Account) (updateAccountErr error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, err := m.db.Exec("UPDATE accounts SET properties=?,publicprops=?,modifiedAt=? WHERE uid=?;",
		userAccount.PropertiesData,
		userAccount.PublicPropsData,
		time.Now(),
		userAccount.UID); err != nil {
		log.Printf("[sqlite3][update-properties] error: %v \r\n", err)
		updateAccountErr = err
		return
	}
	updateAccountErr = nil
	return
}

func (m *Module) updateAccount(userAccount *Account) (updateAccountErr error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, err := m.db.Exec("UPDATE accounts SET roles=?,properties=?,publicprops=?,modifiedAt=? WHERE uid=?;",
		userAccount.Roles,
		userAccount.PropertiesData,
		userAccount.PublicPropsData,
		time.Now(),
		userAccount.UID); err != nil {
		log.Printf("[sqlite3][update-account] error: %v \r\n", err)
		updateAccountErr = err
		return
	}
	updateAccountErr = nil
	return
}

func (m *Module) updatePassword(userAccount *Account) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, err := m.db.Exec("UPDATE accounts SET modifiedAt=?,hash=?,failCount=0,locked=0 WHERE uid=?;",
		time.Now(),
		userAccount.HashedPassword,
		userAccount.UID); err != nil {
		log.Printf("[sqlite3][changePassword] error: %v \r\n", err)
		return err
	}
	return nil
}

func (m *Module) unlockAccount(userAccount *Account) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, err := m.db.Exec("UPDATE accounts SET modifiedAt=?,failCount=0,locked=0 WHERE uid=?;",
		time.Now(),
		userAccount.UID); err != nil {
		log.Printf("[sqlite3][unlockAccount] error: %v \r\n", err)
		return err
	}
	return nil
}

//resetPassword - by email
func (m *Module) resetPassword(a *Account, email string) error {
	newPassword := utils.RandomString(8) + "!0"
	if err := a.setPassword(newPassword); err != nil {
		return err
	}
	if err := m.updatePassword(a); err != nil {
		return err
	}
	body := "您的新密码是：" + newPassword
	return sendMail(email, body)
}
