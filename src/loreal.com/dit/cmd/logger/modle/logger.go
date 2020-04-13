package modle

import (
	"encoding/json"
	"log"
	"time"
)

// Level logger level
type Level string

const (
	// LevelInfo INFO
	LevelInfo Level = "INFO"
	// LevelTrace TRACE
	LevelTrace Level = "TRACE"
	// LevelWarning WARNING
	LevelWarning Level = "WARNING"
	// LevelError ERROR
	LevelError Level = "ERROR"
)

// Logger record log
type Logger struct {
	Project   string      `bson:"project"    json:"project"`    // 项目名称
	Method    string      `bson:"method"     json:"method"`     // 请求方法
	Path      string      `bson:"path"       json:"path"`       // 打印的日志路径
	Level     Level       `bson:"level"      json:"level"`      // 日志级别
	Content   interface{} `bson:"content"    json:"content"`    // 日志内容
	CreatedAt time.Time   `bson:"created_at" json:"created_at"` // 日志创建时间
}

// Contents implement notification interface
func (l *Logger) Contents() []byte {
	b, err := json.Marshal(l)
	if err != nil {
		log.Println(err.Error())
	}
	return b
}

// AccountID implement notification interface
func (l *Logger) AccountID() string {
	return l.Project
}
