package main

import "fmt"

//ErrMissingRuntime - cannot found runtime by name
var ErrMissingRuntime = fmt.Errorf("missing runtime")

//ErrUnfollow - 用户未关注
var ErrUnfollow = fmt.Errorf("用户未关注")

//ErrMsgFailed -消息发送失败
var ErrMsgFailed = fmt.Errorf("消息发送失败")

//MsgState - 消息发送状态
type MsgState int

const (
	//MsgStateNew - Initial state
	MsgStateNew MsgState = iota
	//MsgStateSent - 消息已发送
	MsgStateSent
	//MsgStateUnfollow - 用户未关注
	MsgStateUnfollow
	//MsgStateFailed - 发送失败
	MsgStateFailed
)

func (ms MsgState) String() string {
	switch ms {
	case MsgStateNew:
		return "消息未发送"
	case MsgStateSent:
		return "消息已发送"
	case MsgStateUnfollow:
		return "用户未关注"
	case MsgStateFailed:
		return "发送失败"
	default:
		return ""
	}
}
