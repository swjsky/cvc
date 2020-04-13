package resource

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/microcosm-cc/bluemonday"

	"loreal.com/dit/module/modules/account"
	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"
	"strconv"
	"time"
)

const fileUploadHTML = `
<html>
<head>
<meta http-equiv="content-type" content="text/html;charset=utf-8">
</head>
<body>
<form enctype="multipart/form-data" action="./resource" method="POST">
    Files to send: <input type="file" name="files[]" multiple="multiple" /><br>
    Resource Type: <input type="text" name="rtype" value="" /><br>
    Description: <input type="text" name="desc" value="" /><br>
    Grant access to roles: <input type="text" name="grant" value="" /><br>
    Expires in hours: <input type="text" name="expire-hours" value="" /><br>
    <input type="submit" value="Send Files" />
</form>
</body>
</html>
`

var sanitizePolicy *bluemonday.Policy

func init() {
	sanitizePolicy = bluemonday.UGCPolicy()
}

func getMultiPartFiles(r *http.Request, key string) ([]*multipart.FileHeader, error) {
	if r.MultipartForm == nil {
		err := r.ParseMultipartForm(8 << 20)
		if err != nil {
			return nil, err
		}
	}
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if fhs := r.MultipartForm.File[key]; len(fhs) > 0 {
			return fhs, nil
		}
	}
	return nil, http.ErrMissingFile
}

func (m *Module) upload(caller *account.Account, w http.ResponseWriter, r *http.Request) {
	fileInfos, err := getMultiPartFiles(r, "files[]")
	if err != nil {
		http.Error(w, "Missing Files", 500)
		return
	}
	description := sanitizePolicy.Sanitize(r.FormValue("desc"))
	rtype := sanitizePolicy.Sanitize(r.FormValue("rtype"))
	grant := sanitizePolicy.Sanitize(r.FormValue("grant"))
	if grant == "" {
		grant = "user,anonymous"
	}
	expireHours := -1
	expireHoursStr := sanitizePolicy.Sanitize(r.FormValue("expire-hours"))
	if expireHoursInt, err := strconv.Atoi(expireHoursStr); err == nil {
		expireHours = expireHoursInt
	}

	resources := make([]*Resource, 0, len(fileInfos))

	for _, fi := range fileInfos {
		ext := path.Ext(fi.Filename)
		if sourceFile, err := fi.Open(); err == nil {
			defer sourceFile.Close()
			rid := m.newRID()
			path := m.prepareUploadFolder(caller.UID)
			fileName := rid.Hex() + ext
			f, err := os.Create(path + fileName)
			if err != nil {
				http.Error(w, "Server Error", 500)
				return
			}
			defer f.Close()
			if size, err := io.Copy(f, sourceFile); err == nil {
				r := &Resource{
					ID:           rid,
					Owner:        caller.UID,
					OriginalName: fi.Filename,
					Ext:          ext,
					Description:  description,
					Type:         rtype,
					GrantedRoles: grant,
					Size:         size,
					Mime:         fi.Header.Get("Content-Type"),
					Upload:       time.Now(),
					Expires:      time.Unix(0, 0),
				}
				if expireHours > 0 {
					r.Expires = time.Now().Add(time.Duration(expireHours) * time.Hour)
				}

				if err := m.append(r); err == nil {
					resources = append(resources, r)
				} else {
					http.Error(w, "Receive Error", http.StatusResetContent)
					f.Close()
					os.Remove(path + fileName)
					return
				}
			} else {
				f.Close()
				os.Remove(path + fileName)
				http.Error(w, "Data Error", http.StatusResetContent)
				return
			}
		} else {
			http.Error(w, "Send Error", http.StatusResetContent)
			return
		}
	}
	//fmt.Fprintf(w, "上传文件的大小为: %d", file.(Sizer).Size())
	//w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Header().Set("Content-type", "application/json;charset=utf-8")
	b, _ := json.MarshalIndent(resources, "", "  ")
	w.Write(b)
	return
}

func (m *Module) download(caller *account.Account, rid, options string, w http.ResponseWriter, r *http.Request) {
	if caller == nil {
		caller = &account.Account{
			UID:   "anonymous",
			Roles: "anonymous",
		}
	}
	if resource, err := m.get(rid); err == nil {
		if caller.IsSelfOrAdmin(resource.Owner) || caller.IsInRole(resource.GrantedRoles) {
			keepname := false
			if options != "" && options == "keepname" {
				keepname = true
				options = ""
			}

			w.Header().Set("X-Content-Type-Options", "nosniff")
			if DefaultConfig.MimeAccepted(resource.Mime) {
				w.Header().Set("Content-type", resource.Mime)
			} else {
				w.Header().Set("Content-type", "application/octet-stream")
				if keepname {
					w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s";`, resource.OriginalName))
				} else {
					w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.%s";`, resource.ID.Hex(), resource.Ext))
				}
			}
			originFileName := resource.FullPath(m.UploadPath, "")
			targetFileName := resource.FullPath(m.UploadPath, options)
			if _, err := os.Stat(targetFileName); err == nil {
				http.ServeFile(w, r, targetFileName)
				return
			}
			var handled, hasHandlerErr bool
			for _, h := range m.MimeHandlers {
				if h.Match(resource.Mime) {
					handled = true
					if err := h.Process(originFileName, targetFileName, resource.Mime, options); err != nil {
						log.Println("[ERR][Handler]", err)
						hasHandlerErr = true
					}
				}
			}
			if handled && !hasHandlerErr {
				http.ServeFile(w, r, targetFileName)
				return
			}
			http.ServeFile(w, r, originFileName)
			return
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	http.Error(w, "404", http.StatusNotFound)
	return

}

func (m *Module) registerEndpoints(u middlewares.RoleVerifier) {

	m.MountingPoints[""] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			caller := &account.Account{
				UID:   r.Header.Get("uid"),
				Roles: r.Header.Get("roles"),
			}

			if caller == nil {
				http.Error(w, "500", http.StatusInternalServerError)
				return
			}

			switch r.Method {
			case "POST":
				m.upload(caller, w, r)
				return
			case "GET":
				q := r.URL.Query()
				rid := q.Get("rid")
				size := q.Get("size")
				keepname := q.Get("keepname")
				if rid == "" {
					// 上传页面
					w.Header().Add("Content-Type", "text/html")
					w.WriteHeader(200)
					io.WriteString(w, fileUploadHTML)
					return
				}
				//资源下载
				var options string
				if keepname != "" {
					options = "keepname"
				} else {
					options = size
				}

				m.download(caller, rid, options, w, r)
			}
		}),
		middlewares.ServerInstrumentation("resource", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.CORS("*", "POST", "Content-Type, Accept, Authorization", ""),
		middlewares.BasicAuthOrTokenAuthWithRole(u, "resource", DefaultConfig.UploadRoles),
	)

	m.MountingPoints["public"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "POST":
				http.Error(w, "Not Support", http.StatusNotAcceptable)
				return
			case "GET":
				q := r.URL.Query()
				rid := q.Get("rid")
				size := q.Get("size")
				if size == "" {
					size = q.Get("w")
				}
				if rid == "" {
					http.Error(w, "404", http.StatusNotFound)
					return
				}
				if size == "" && strings.Contains(rid, `\u0026`) {
					params, err := url.ParseQuery("rid=" + strings.Replace(rid, `\u0026`, `&`, -1))
					if err != nil {
						http.Error(w, "500", http.StatusInternalServerError)
						return
					}
					rid = params.Get("rid")
					size = params.Get("size")
					if size == "" {
						size = params.Get("w")
					}
				}
				//资源下载
				m.download(nil, rid, size, w, r)
			}
		}),
		middlewares.ServerInstrumentation("get-public", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
	)

	m.MountingPoints["cleanup"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method Not Acceptable", http.StatusNotAcceptable)
				return
			}
			if err := m.removeExpires(); err != nil {
				http.Error(w, "500", http.StatusInternalServerError)
				return
			}
			w.Write([]byte("OK"))
		}),
		middlewares.ServerInstrumentation("cleanup", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BasicAuthOrTokenAuthWithRole(u, "resource", "admin"),
	)

}
