package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//debugHandler - vehicle plate management
//endpoint: debug
//method: GET
func (a *App) debugHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		q := r.URL.Query()
		module := sanitizePolicy.Sanitize(q.Get("module"))
		if module == "" {
			module = "*"
		}
		DEBUG = "" != sanitizePolicy.Sanitize(q.Get("debug"))
		INFOLEVEL, _ = strconv.Atoi(sanitizePolicy.Sanitize(q.Get("level")))
		LOGLEVEL, _ = strconv.Atoi(sanitizePolicy.Sanitize(q.Get("log-level")))
		var result struct {
			Module    string `json:"module"`
			DebugFlag bool   `json:"debug-flag"`
			InfoLevel int    `json:"info-level"`
			LogLevel  int    `json:"log-level"`
		}
		switch module {
		case "*":
		}
		result.Module = module
		result.DebugFlag = DEBUG
		result.InfoLevel = INFOLEVEL
		result.LogLevel = LOGLEVEL
		outputJSON(w, result)
	default:
		outputJSON(w, map[string]interface{}{
			"errcode": -100,
			"errmsg":  "Method not acceptable",
		})
	}
}

//feUpgradeHandler - upgrade fe
//endpoint: maintenance/fe/upgrade
//method: GET
func (a *App) feUpgradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}
	q := r.URL.Query()
	_ = q
	username := sanitizePolicy.Sanitize(q.Get("u"))
	password := q.Get("p")
	q.Encode()
	var resp struct {
		ErrCode    int    `json:"errcode,omitempty"`
		ErrMessage string `json:"errmsg,omitempty"`
		Data       string `json:"data,omitempty"`
	}
	const feFolder = "./fe/"
	var err error
	var strPullURL string
	if strPullURL, err = getGitURL(feFolder); err != nil {
		log.Println(err)
	}
	log.Println(strPullURL)
	if strPullURL == "" {
		resp.Data = "empty pull url"
		outputJSON(w, resp)
		return
	}
	buffer := bytes.NewBuffer(nil)
	if pullURL, err := url.Parse(strPullURL); err == nil {
		pullURL.User = url.UserPassword(username, password)
		strPullURL = pullURL.String()
		resp.Data += pullURL.String()
	}

	if err := runShellCmd(nil, buffer, buffer, feFolder, "git", "pull", strPullURL); err != nil {
		logError(err, "git pull")
	}
	if err := runShellCmd(nil, buffer, buffer, feFolder, "npm", "run", "build"); err != nil {
		logError(err, "git pull")
	}
	resp.Data = buffer.String()
	outputText(w, buffer.Bytes())
}

func runShellCmd(stdin io.Reader, stdout, stderr io.Writer, workingDir, cmd string, args ...string) error {
	di, err := os.Stat(workingDir)
	if err != nil {
		return err
	}
	if !di.IsDir() {
		return fmt.Errorf("invalid working dir")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = stderr
	c.Dir = workingDir
	if err := c.Run(); err != nil {
		log.Println("[ERR] - run", err)
	}
	return nil
}

func getGitURL(path string) (string, error) {
	fi, err := os.Stat(path + ".git/")
	if os.IsNotExist(err) || !fi.IsDir() {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "--push", "origin")
	cmd.Dir = path
	rc, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		log.Println(err)
		return "", err
	}
	go func(cancel context.CancelFunc) {
		if err := cmd.Run(); err != nil {
			log.Println("[ERR] - run", err)
		}
		defer cancel()
	}(cancel)

	data, err := ioutil.ReadAll(rc)
	if err := rc.Close(); err != nil {
		return strings.TrimRight(string(data), "\r\n"), err
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}
