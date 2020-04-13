package task

import (
	"errors"
	"fmt"
	"log"
	"sync"

	uuid "github.com/satori/go.uuid"
)

//Manager - background tasks
type Manager struct {
	Owner       interface{}               `json:"-"`
	Instances   map[string]*Task          `json:"tasks,omitempty"`
	RunningPool chan int                  `json:"-"`
	Registry    map[string]*RegistryEntry `json:"-"`
	Mutex       *sync.RWMutex             `json:"-"`
}

//NewManager - constractor for manager
func NewManager(owner interface{}, maxRunningTasks int) *Manager {
	m := &Manager{
		Owner:       owner,
		Instances:   make(map[string]*Task, 0),
		RunningPool: nil,
		Registry:    make(map[string]*RegistryEntry, 1),
		Mutex:       &sync.RWMutex{},
	}
	if maxRunningTasks >= 0 {
		m.RunningPool = make(chan int, maxRunningTasks)
	}
	return m
}

//Register - create a new background task instance
func (m *Manager) Register(name string, handler func(task *Task, args ...string), maxInstances int32) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Registry[name] = &RegistryEntry{
		Name:         name,
		Handler:      handler,
		MaxInstances: maxInstances,
	}
}

//RegisterWithContext - create a new background task instance
func (m *Manager) RegisterWithContext(name, context string, handler func(task *Task, args ...string), maxInstances int32) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Registry[name] = &RegistryEntry{
		Name:         name,
		Context:      context,
		Handler:      handler,
		MaxInstances: maxInstances,
	}
}

//NewInstance - create a new background task instance
func (m *Manager) NewInstance(name string) *Task {
	m.Mutex.RLock()
	re, ok := m.Registry[name]
	m.Mutex.RUnlock()
	if !ok {
		log.Println("[ERR] - Unregisted Task:", name)
		return nil
	}
	if !re.GetToken() {
		log.Printf("[WARN] - Max instance reached for task: [%s]\r\n", name)
		return nil
	}
	inst :=
		&Task{
			ID:      fmt.Sprintf("%x", uuid.NewV4().Bytes()),
			Owner:   m,
			Name:    name,
			Context: re.Context,
			Handler: re.Handler,
			Metrics: []*Metrics{NewMetrics("", -1)},
			mutex:   &sync.RWMutex{},
		}
	inst.SetState(StateNew)
	m.Mutex.Lock()
	m.Instances[inst.ID] = inst
	m.Mutex.Unlock()
	return inst
}

//GetInstance - find task by name
func (m *Manager) GetInstance(ID string) *Task {
	m.Mutex.RLock()
	task, ok := m.Instances[ID]
	m.Mutex.RUnlock()
	if !ok {
		return nil
	}
	return task
}

//RunTask - by name
func (m *Manager) RunTask(name string, args ...string) string {
	inst := m.NewInstance(name)
	if inst == nil {
		return ""
	}
	if m.RunningPool != nil {
		m.RunningPool <- 1
	}

	if wg := inst.Run(args...); wg != nil {
		//task finishing procedure
		go func(t *Task, wg *sync.WaitGroup) {
			wg.Wait()
			//Remove from RunningTasks
			inst.Owner.Mutex.Lock()
			//delete(inst.Owner.Instances, inst.ID)
			inst.Owner.Registry[inst.Name].ReleaseToken()
			inst.Owner.Mutex.Unlock()
			//Release global running pool
			if inst.Owner.RunningPool != nil {
				<-inst.Owner.RunningPool
			}
			if INFOLEVEL >= 3 {
				log.Printf("[INFO] - Task [%s][%s] Disposed.\r\n", inst.Name, inst.ID)
			}
		}(inst, wg)
		return inst.ID
	}
	return ""
}

//Send - send control command to task by ID
func (m *Manager) Send(ID, ctrlCommand string) error {
	m.Mutex.RLock()
	task, ok := m.Instances[ID]
	m.Mutex.RUnlock()
	if !ok {
		return errors.New("Task not found")
	}
	go func(t *Task, cmd string) {
		log.Printf("[INFO] - Task:[%s] Sending control command '%s'...\r\n", t.Name, cmd)
		if t.ControlBus != nil {
			t.ControlBus <- cmd
		}
		log.Printf("[INFO] - Task:[%s] Control command '%s' received\r\n", t.Name, cmd)
	}(task, ctrlCommand)
	return nil
}

//SendAll - send control command to all task instances
func (m *Manager) SendAll(ctrlCommand string) error {
	go func(m *Manager, cmd string) {
		m.Mutex.RLock()
		defer m.Mutex.RUnlock()
		for _, inst := range m.Instances {
			if inst.State != StateRunning {
				continue
			}
			if inst.ControlBus == nil {
				continue
			}
			log.Printf("[INFO] - Task:[%s] Sending control command '%s'...\r\n", inst.Name, cmd)
			inst.ControlBus <- cmd
			log.Printf("[INFO] - Task:[%s] Control command '%s' received\r\n", inst.Name, cmd)
		}
	}(m, ctrlCommand)
	return nil
}

//Clear - remove all finished task instances
func (m *Manager) Clear() error {
	keysToRemove := make([]string, 0, 0)
	m.Mutex.RLock()
	for key, inst := range m.Instances {
		switch inst.State {
		case StateCompleted, StateCanceled:
			keysToRemove = append(keysToRemove, key)
		}
	}
	m.Mutex.RUnlock()
	m.Mutex.Lock()
	for _, key := range keysToRemove {
		delete(m.Instances, key)
	}
	m.Mutex.Unlock()
	return nil
}
