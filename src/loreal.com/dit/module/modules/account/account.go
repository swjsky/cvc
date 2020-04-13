//Package account - user account service for CEH applications
package account

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//Account - data struct to store user account info.
type Account struct {
	UID             string      `json:"uid"`
	HashedPassword  []byte      `json:",omitempty"`
	Roles           string      `json:"roles"`
	Properties      interface{} `json:"properties,omitempty"`
	PropertiesData  []byte      `json:"-"`
	PublicProps     interface{} `json:"publicprops,omitempty"`
	PublicPropsData []byte      `json:"-"`
	CreateAt        time.Time   `json:"create,omitempty"`
	ModifiedAt      time.Time   `json:"modified,omitempty"`
	LoginAt         time.Time   `json:"login,omitempty"`
	ExpiresOn       time.Time   `json:"expires,omitempty"`
	Enabled         int         `json:"enabled"`
	Locked          int64       `json:"locked"`
}

//NewAccount - create user account from uid & password
func NewAccount(uid, password, roles, properties string) *Account {
	result := &Account{
		UID: uid,
	}
	if err := result.setPassword(password); err != nil {
		return nil
	}
	return result
}

func (a *Account) setPassword(password string) error {
	if DefaultAccountConfig.PasswordRuleRegexp != nil {
		match, err := DefaultAccountConfig.PasswordRuleRegexp.MatchString(password)
		if err != nil {
			log.Println("[ERR] - Violation of password rule, uid:", a.UID, err)
			return ErrPasswordRule
		}
		if !match {
			log.Println("[WARN] - Violation of password rule, uid:", a.UID)
			return ErrPasswordRule
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("[ERR] - error when hash password, uid:", a.UID, err)
		return ErrPasswordRule
	}
	a.HashedPassword = hashedPassword
	return nil
}

//VerifyPassword - Verify Password
func (a *Account) VerifyPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(a.HashedPassword, []byte(password)) == nil
}

//SetPassword - Set Password
func (a *Account) SetPassword(newPass1, newPass2 string) error {
	if newPass1 != newPass2 {
		log.Println("[WARN] - [cpw] password not match")
		return ErrPasswordNotMatch
	}
	return a.setPassword(newPass1)
}

//IsInRole - Check if the account is in role
func (a *Account) IsInRole(role string) bool {
	targetRoles := strings.Split(role, ",")
	roles := strings.Split(a.Roles, ",")
	for _, r := range roles {
		for _, tr := range targetRoles {
			if r == tr {
				return true
			}
		}
	}
	return false
}

//IsAdmin - Check if the account is admin
func (a *Account) IsAdmin() bool {
	return a.IsInRole("admin")
}

//IsSelfOrAdmin - Check if the account is admin or himself
func (a *Account) IsSelfOrAdmin(selfUID string) bool {
	return selfUID != "" && selfUID == a.UID || a.IsInRole("admin")
}

//IsAnotherAdmin - Check if the account is another admin but not himself
func (a *Account) IsAnotherAdmin(selfUID string) bool {
	return selfUID != a.UID && a.IsInRole("admin")
}

func mergeProperties(source, target interface{}) interface{} {
	//Can migrate key values?
	if sourceObj, typeOk := source.(map[string]interface{}); typeOk {
		if targetObj, typeOk := target.(map[string]interface{}); typeOk {
			//Migrate key values
			for key, sourceValue := range sourceObj {
				if sourceValueObj, isMap := sourceValue.(map[string]interface{}); isMap {
					targetObj[key] = mergeProperties(sourceValueObj, targetObj[key])
				} else {
					targetObj[key] = sourceValue
				}
			}
			return targetObj
		}
		return sourceObj
	}
	return source
}

func marshalProperties(target interface{}) []byte {
	result, err := json.Marshal(target)
	if err != nil {
		log.Println("[ERR][UserAccount.mergeProperties]:", err)
	}
	return result
}

func unmarshalProperties(data []byte) (interface{}, error) {
	if data == nil {
		return nil, nil
	}
	result := make(map[string]interface{})
	if err := json.Unmarshal(data, &result); err != nil {
		log.Println("[ERR][UserAccount.unmarshalProperties]:", err)
		return nil, err
	}
	return result, nil
}

//SetProperties - into account
func (a *Account) SetProperties(props interface{}) (err error) {
	a.Properties = mergeProperties(props, a.Properties)
	a.PropertiesData = marshalProperties(a.Properties)
	return
}

//GetPropertie - into account
func (a *Account) GetPropertie(key string) interface{} {
	if obj, ok := a.PublicProps.(map[string]interface{}); ok {
		return obj[key]
	}
	return nil
}

//SetPublicProps - into account
func (a *Account) SetPublicProps(props interface{}) (err error) {
	a.PublicProps = mergeProperties(props, a.PublicProps)
	a.PublicPropsData = marshalProperties(a.PublicProps)
	return
}

//ParseProperties - parse propertiesData into properties
func (a *Account) ParseProperties() (err error) {
	a.Properties, err = unmarshalProperties(a.PropertiesData)
	return
}

//ParsePublicProps - parse propertiesData into properties
func (a *Account) ParsePublicProps() (err error) {
	a.PublicProps, err = unmarshalProperties(a.PublicPropsData)
	return
}

//Parse Properties and PublicProps
func (a *Account) Parse() (err error) {
	err1 := a.ParseProperties()
	err2 := a.ParsePublicProps()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return
}

//AsString - convert interface{} to string
func AsString(value interface{}) string {
	if v, ok := value.(string); ok {
		return v
	}
	return ""
}

//GetPropertie - into account
func GetPropertie(source interface{}, key string) interface{} {
	if obj, ok := source.(map[string]interface{}); ok {
		return obj[key]
	}
	return nil
}
