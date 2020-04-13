package account

import (
	"fmt"
)

//ErrWebTokenExpired - Web token expired
var ErrWebTokenExpired = fmt.Errorf("Web token expired")

//WebTokenValid - parse token and check whether token is valid
//Implements middlewares.WebTokenVerifier interface
func (m *Module) WebTokenValid(
	webToken string,
	uid *string,
	roles *string,
	properties *[]byte,
	publicprops *[]byte,
) bool {
	u := m.VerifyToken(webToken)
	if u == nil {
		return false
	}
	*uid = u.UID
	*roles = u.Roles
	*properties = u.PropertiesData
	*publicprops = u.PublicPropsData
	return *uid != ""
}

//GetWebTokenCookieName - Implements middlewares.WebTokenVerifier interface
func (m *Module) GetWebTokenCookieName() string {
	return DefaultAccountConfig.CookieName
}
