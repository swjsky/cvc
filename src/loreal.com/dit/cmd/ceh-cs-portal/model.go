package main

//APIStatus - general api result
type APIStatus struct {
	ErrCode    int    `json:"code,omitempty"`
	ErrMessage string `json:"msg,omitempty"`
}

//GatewayRequest - requst encoding struct for gateway
type GatewayRequest struct {
	Path    string                 `json:"path,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}
