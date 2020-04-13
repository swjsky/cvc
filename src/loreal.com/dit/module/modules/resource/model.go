//Package resource - implements the resource upload/donwload service
package resource

import "gopkg.in/mgo.v2/bson"
import "time"

//Resource struct for CEH
type Resource struct {
	ID           bson.ObjectId `bson:"_id"`
	Owner        string        `json:"owner"`
	Type         string        `json:"type"`
	Description  string        `json:"description"`
	GrantedRoles string        `json:"granted-roles"`
	OriginalName string        `json:"original-name"`
	Ext          string        `json:"ext"`
	Size         int64         `json:"size"`
	Mime         string        `json:"mime"`
	Upload       time.Time     `json:"upload"`
	Expires      time.Time     `json:"expires"`
}

//FullPath - returns full url to resource files
func (r *Resource) FullPath(uploadPath string, suffix string) string {
	if suffix != "" {
		return uploadPath + "/" + r.Owner + "/" + r.ID.Hex() + "_" + suffix + r.Ext
	}
	return uploadPath + "/" + r.Owner + "/" + r.ID.Hex() + r.Ext
}

//MimeHandler - process resource base on Mime type
type MimeHandler interface {
	Match(mime string) bool
	Process(sourceFile, targetFile, mime, options string) error
}
