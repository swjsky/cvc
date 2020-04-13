package task

import (
	"fmt"
	"sync"
	"time"
)

//Metrics - task metrics for reporting
type Metrics struct {
	Stage   string         `json:"stage,omitempty"`
	Data    map[string]int `json:"data,omitempty"`
	Status  string         `json:"status,omitempty"`
	Logs    []string       `json:"logs,omitempty"`
	MaxLogs int            `json:"-"`
	mutex   *sync.RWMutex
}

//NewMetrics - constractor
func NewMetrics(stage string, maxLogs int) *Metrics {
	return &Metrics{
		Stage:   stage,
		Data:    make(map[string]int, 0),
		MaxLogs: maxLogs,
		Logs:    make([]string, 0, 10),
		mutex:   &sync.RWMutex{},
	}
}

//Init - m.Data["progress"] & m.Data["total"]
func (m *Metrics) Init(values ...struct {
	Key   string
	Value int
}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, item := range values {
		m.Data[item.Key] = item.Value
	}
}

//Get - m.Data[key]
func (m *Metrics) Get(key string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Data[key]
}

//Set - m.Data[key] to value
func (m *Metrics) Set(key string, value int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Data[key] = value
}

//Log - log recent messages.
func (m *Metrics) Log(format string, args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006/01/02 15:04:05")}, args...)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Logs = append(m.Logs, fmt.Sprintf("%s "+format, args...))
	if m.MaxLogs > 0 && len(m.Logs) >= m.MaxLogs {
		m.Logs = m.Logs[1:]
	}
}

//Step - m.Data["progress"]
func (m *Metrics) Step(key string, delta int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Data[key] += delta
}

//StepOne - m.Data["progress"]
func (m *Metrics) StepOne(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Data[key]++
}
