package coupon

// MessageType 消息类型
type MessageType int32

const (
	//MTIssue - Issue
	MTIssue MessageType = iota
	//MTRevoked - Revoked
	MTRevoked
	//MTRedeemed - Redeemed
	MTRedeemed
	//MTUnknown - Unknown
	MTUnknown
)

// Message 消息结构
type Message struct {
	Type    MessageType `json:"type,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}
