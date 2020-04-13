package account

import (
	"encoding/base64"
	"fmt"
	"net/mail"

	"loreal.com/dit/utils/smtp"
)

func sendMail(toEmail, body string) error {
	b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	host := DefaultAccountConfig.SMTPServer
	email := DefaultAccountConfig.SMTPUser
	password := DefaultAccountConfig.SMTPPassword
	//toEmail := "bin.hu@loreal.com"
	from := mail.Address{
		Name:    "noreply",
		Address: email,
	}
	to := mail.Address{
		Name:    "接收人",
		Address: toEmail,
	}
	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", b64.EncodeToString([]byte("密码重置")))
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"
	header["Content-Transfer-Encoding"] = "base64"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + b64.EncodeToString([]byte(body))
	auth := smtp.PlainAuth(
		"",
		email,
		password,
		host,
	)
	return smtp.SendMail(
		host+":25",
		auth,
		email,
		[]string{to.Address},
		[]byte(message),
	)
}
