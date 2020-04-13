package account

//TokenPackage - struct to send token over HTTP API
type TokenPackage struct {
	Token       string      `json:"token"`
	ExpiresIn   int         `json:"expires_in"` /*Expires in seconds*/
	Roles       string      `json:"roles,omitempty"`
	Properties  interface{} `json:"properties,omitempty"`
	PublicProps interface{} `json:"publicprops,omitempty"`
}
