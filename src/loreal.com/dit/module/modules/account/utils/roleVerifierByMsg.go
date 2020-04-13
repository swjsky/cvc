//Package utils implements TokenRoleVerifier interface by send/receive messages
package utils

import (
	"loreal.com/dit/middlewares"
	"loreal.com/dit/module"
	"loreal.com/dit/module/modules/account"
)

//RoleVerifierByMessage - implements TokenRoleVerifier interface by sending message to base module
type RoleVerifierByMessage struct {
	Base *module.Module
}

//VerifyTokenAndRole - implements TokenRoleVerifier interface
func (uv RoleVerifierByMessage) VerifyTokenAndRole(token, realm, role string, callback middlewares.VerifyTokenCallback) {
	resultChan := make(chan interface{}, 0)
	uv.Base.SendWithChan("getUserAccountByToken", resultChan, token)
	if result, ok := <-resultChan; ok {
		if userAccount, isUserAccount := result.(*account.Account); isUserAccount {
			if userAccount != nil {
				callback(userAccount.IsInRole(role), userAccount.UID, userAccount.Roles)
			} else {
				callback(false, "", "")
			}
		}
	}
}

//VerifyPasswordAndRole - implements TokenRoleVerifier interface
func (uv RoleVerifierByMessage) VerifyPasswordAndRole(uid, password, realm, role string, callback middlewares.VerifyTokenCallback) {
	resultChan := make(chan interface{}, 0)
	uv.Base.SendWithChan("getUserAccountByPassword", resultChan, uid, password)
	if result, ok := <-resultChan; ok {
		if userAccount, isUserAccount := result.(*account.Account); isUserAccount {
			if userAccount != nil {
				callback(userAccount.IsInRole(role), userAccount.UID, userAccount.Roles)
			} else {
				callback(false, "", "")
			}
		}
	}
}
