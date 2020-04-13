package account

import "loreal.com/dit/middlewares"

//VerifyTokenAndRole - implements TokenRoleVerifier interface
func (m *Module) VerifyTokenAndRole(token, realm, role string, callback middlewares.VerifyTokenCallback) {
	userAccount := m.VerifyToken(token)
	if userAccount != nil {
		callback(userAccount.IsInRole(role), userAccount.UID, userAccount.Roles)
	} else {
		callback(false, "", "")
	}
}

//VerifyPasswordAndRole - implements TokenRoleVerifier interface
func (m *Module) VerifyPasswordAndRole(uid, password, realm, role string, callback middlewares.VerifyTokenCallback) {
	userAccount, _ := m.Authenticate(uid, password)
	if userAccount != nil {
		callback(userAccount.IsInRole(role), userAccount.UID, userAccount.Roles)
	} else {
		callback(false, "", "")
	}
}
