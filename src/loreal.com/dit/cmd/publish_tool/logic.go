package main

import (
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

func (m *Module) publish(pn, branch, bn, basedir string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sh_name := "publish.sh"
	if basedir == "src" {
		sh_name = "publish4all.sh"
	}
	result, err := m.runCommand("./"+sh_name, branch, pn, bn)
	if err != nil {
		return "[ERR]: " + err.Error()
	}
	if strings.Contains(result, "build error") {
		return result[strings.Index(result, "build error")+len("build error"):]
	}
	return "success"
}

func (m *Module) runCommand(
	command string,
	args ...string,
) (
	string,
	error,
) {
	cmd := exec.Command(command, args...)
	// 获取输出对象，可以从该对象中读取输出结果
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("[StdoutPipe Err]:", err.Error())
		return "", err
	}
	// 保证关闭输出流
	defer stdout.Close()
	// 运行命令
	if err := cmd.Start(); err != nil {
		log.Println("[Start Err]:", err.Error())
		return "", err
	}
	// 读取输出结果
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Println("[ReadAll Err]:", err.Error())
		return "", err
	}
	return string(opBytes), nil
}

//ReadUser implements UserReader interface
func (m *Module) ReadUser(username, realm string) (pass string, ok bool) {
	if userAccounts[username] != "" {
		return userAccounts[username], true
	}
	return "", false
}
