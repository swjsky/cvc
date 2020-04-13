package account

import (
	"log"

	regexp "github.com/dlclark/regexp2"
	"loreal.com/dit/utils"
)

//Config - for account module
type Config struct {
	CookieName         string         `json:"cookie-name"`
	CookieDomain       string         `json:"cookie-domain"`
	CookiePath         string         `json:"cookie-path"`
	CookieSecure       bool           `json:"cookie-secure"`
	WebTokenKey        string         `json:"web-token-key"`
	PasswordRule       string         `json:"password-rule"`
	SMTPServer         string         `json:"smtp-server"`
	SMTPUser           string         `json:"smtp-user"`
	SMTPPassword       string         `json:"smtp-password"`
	LoginLimit         int            `json:"login-limit"`
	LoginPath          string         `json:"login-path"`
	LogoutPath         string         `json:"logout-path"`
	MaxAge             int            `json:"max-age"`
	PasswordRuleRegexp *regexp.Regexp `json:"-"`
}

//DefaultAccountConfig for account module
var DefaultAccountConfig *Config

//DefaultPasswordRule - 默认密码要求：必须包含大写字母，小写字母，数字，特殊字符四种中的三种，长度要求8到30位
const DefaultPasswordRule = `(?=.*[0-9])(?=.*[a-z])(?=.*[A-Z])(?=.*[^a-zA-Z0-9]).{8,30}`

//GetDefaultConfig for account module
func GetDefaultConfig() *Config {
	DefaultAccountConfig = &Config{
		CookieName:   "dit-web-token",
		CookieDomain: "www.test-domain.com",
		CookiePath:   "/",
		CookieSecure: false,
		WebTokenKey:  "8b543259fec05380ab886d553322c4e1",
		PasswordRule: DefaultPasswordRule,
		SMTPServer:   "139.219.137.62",
		SMTPUser:     "ceh@e-loreal.cn",
		SMTPPassword: "ewCgYGk94E",
		LoginLimit:   2,
		LoginPath:    "/",
		LogoutPath:   "/",
		MaxAge:       1800,
	}
	utils.LoadOrCreateJSON("./config/web-token.json", &DefaultAccountConfig)
	//No password rule, any password is acceptable
	if DefaultAccountConfig.PasswordRule == "" {
		log.Println("[INFO] - Empty password rule")
		DefaultAccountConfig.PasswordRuleRegexp = nil
		return DefaultAccountConfig
	}
	regexp1, err := regexp.Compile(DefaultAccountConfig.PasswordRule, regexp.None)
	if err != nil {
		log.Println("[ERR] - Invalid password rule", err)
		DefaultAccountConfig.PasswordRule = DefaultPasswordRule
		DefaultAccountConfig.PasswordRuleRegexp, _ = regexp.Compile(DefaultPasswordRule, regexp.None)
		return DefaultAccountConfig
	}
	DefaultAccountConfig.PasswordRuleRegexp = regexp1
	return DefaultAccountConfig
}

func init() {
	GetDefaultConfig()
}
