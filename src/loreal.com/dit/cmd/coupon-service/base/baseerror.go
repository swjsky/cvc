package base

import (
	"fmt"
)

// ErrorWithCode 作为统一返回给调用端的错误结构。
// TODO: 确认下面的注释在godoc里面
type ErrorWithCode struct {
	// Code：错误编码, 每个code对应一个业务错误。
	Code		int		
	// Message：错误描述, 可以用于显示给用户。	
	Message		string
}

func (e *ErrorWithCode) Error() string {
	return fmt.Sprintf(
`{
	"error-code" : %d,
	"error-message" : "%s"
}`, e.Code, e.Message)
}

//ErrTokenValidateFailed - 找不到一个规则的校验。
var ErrTokenValidateFailed = ErrorWithCode{
	Code:    1500,
	Message: "validate token failed",
}

//ErrTokenExpired - 找不到一个规则的校验。
var ErrTokenExpired = ErrorWithCode{
	Code:    1501,
	Message: "the token is expired",
}
