package resource

import (
	"strings"

	"loreal.com/dit/utils"
)

//Config - for resource module
type Config struct {
	AcceptedMimes []string `json:"accepted-mimes"`
	UploadRoles   string   `json:"upload-roles"`
}

//MimeAccepted - 判断 MIME 类型是否在可接受列表中
func (c *Config) MimeAccepted(mime string) bool {
	mime = strings.ToLower(mime)
	for _, prefix := range c.AcceptedMimes {
		if strings.HasPrefix(mime, prefix) {
			return true
		}
	}
	return false
}

//DefaultConfig for resource module
var DefaultConfig *Config

//GetDefaultConfig for resource module
func GetDefaultConfig() *Config {
	DefaultConfig = &Config{
		UploadRoles: "admin,user,app",
		AcceptedMimes: []string{
			"image/",
			"audio/",
			"video/",
			"application/pdf",
		},
	}
	utils.LoadOrCreateJSON("./config/resource-config.json", &DefaultConfig)

	return DefaultConfig
}

func init() {
	GetDefaultConfig()
}
