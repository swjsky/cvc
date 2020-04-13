//Package task - task manager for backend long run program logic (go routings)
package task

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

//INFOLEVEL - for debug mode
var INFOLEVEL int

func init() {
	INFOLEVEL = 3
}

//Task - background task
type Task struct {
	ID         string                           `json:"taskid,omitempty"`
	Name       string                           `json:"name,omitempty"`
	Context    string                           `json:"context,omitempty"`
	Owner      *Manager                         `json:"-"`
	Handler    func(task *Task, args ...string) `json:"-"`
	ControlBus chan string                      `json:"-"`
	Metrics    []*Metrics                       `json:"metrics,omitempty"`
	Start      time.Time                        `json:"start,omitempty"`
	End        time.Time                        `json:"end,omitempty"`
	Duration   string                           `json:"duration,omitempty"`
	State      State                            `json:"state,omitempty"`
	mutex      *sync.RWMutex
}

//Canceled - check whether task is canceled
func (t *Task) Canceled(msg string) bool {
	select {
	case cmd := <-t.ControlBus:
		if cmd == "stop" {
			log.Printf("[INFO] - Task:[%s] Canceled! %s\r\n", t.Name, msg)
			return true
		}
	default:
	}
	return false
}

//SetState - set state for task instance
func (t *Task) SetState(s State) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	switch s {
	case StateRunning:
		switch t.State {
		case StateNew, StatePaused:
			t.State = StateRunning
			if t.Start.IsZero() {
				t.Start = time.Now()
			}
		}
	case StatePaused:
		if t.State == StateRunning {
			t.State = StatePaused
			t.End = time.Now()
			t.Duration = t.End.Sub(t.Start).String()
		}
	case StateCompleted:
		switch t.State {
		case StateRunning:
			t.State = StateCompleted
			t.End = time.Now()
			t.Duration = t.End.Sub(t.Start).String()
		}
	case StateCanceled:
		switch t.State {
		case StateRunning:
			t.State = StateCanceled
			t.End = time.Now()
			t.Duration = t.End.Sub(t.Start).String()
		}
	}
}

//GetStatus - get status for task instance
func (t *Task) GetStatus() State {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.State
}

//Run - run task by args, return whether ok
func (t *Task) Run(args ...string) *sync.WaitGroup {
	t.mutex.RLock()
	if t.Handler == nil {
		t.mutex.RUnlock()
		log.Printf("[INFO] - Task:[%s] Invalid handler!\r\n", t.Name)
		return nil
	}
	t.mutex.RUnlock()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(task *Task, wg *sync.WaitGroup, args ...string) {
		defer wg.Done()
		t.SetState(StateRunning)
		if INFOLEVEL >= 3 {
			log.Printf("[INFO] - Task:[%s][%s] Running with args: %v ...\r\n", t.Name, t.ID, args)
		}

		//Run task
		t.ControlBus = make(chan string, 0)
		t.Handler(task, args...)
		close(t.ControlBus)
		t.ControlBus = nil
		if INFOLEVEL >= 3 {
			log.Printf("[INFO] - Task:[%s] Finished.\r\n", t.Name)
		}
		t.SetState(StateCompleted)
	}(t, wg, args...)
	return wg
}

//NextStage - advance to next stage
func (t *Task) NextStage(name string, args ...interface{}) {
	t.mutex.Lock()
	t.Metrics = append([]*Metrics{NewMetrics(fmt.Sprintf(name, args...), -1)}, t.Metrics...)
	t.mutex.Unlock()
}

//Log - log message to current metrics
func (t *Task) Log(format string, args ...interface{}) {
	t.mutex.Lock()
	t.Metrics[0].Log(format, args...)
	t.mutex.Unlock()
}

//CurrentMetrics - current metrics
func (t *Task) CurrentMetrics() *Metrics {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.Metrics[0]
}

//StepOne - data[key]++
func (t *Task) StepOne(key string) {
	t.mutex.Lock()
	t.Metrics[0].StepOne(key)
	t.mutex.Unlock()
}

//Step - data[key]+delta
func (t *Task) Step(key string, delta int) {
	t.mutex.Lock()
	t.Metrics[0].Step(key, delta)
	t.mutex.Unlock()
}

//StrNormalize - Normalize string to lower trimed version
func StrNormalize(value *string) {
	*value = strings.ToLower(strings.TrimSpace(*value))
}

//GetArgs - get args by index
func GetArgs(args []string, index int) string {
	if len(args) < index+1 {
		log.Println("[ERR] - [taskHandler] Missing Args")
		return ""
	}
	return args[index]
}
