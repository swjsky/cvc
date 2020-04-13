package task

import (
	"sync/atomic"
)

//RegistryEntry - data record in task mamaner registry
type RegistryEntry struct {
	Name             string
	Context          string
	Handler          func(task *Task, args ...string)
	RunningInstances int32
	MaxInstances     int32
}

//GetToken - get instance token to start an instance
func (re *RegistryEntry) GetToken() bool {
	if atomic.LoadInt32(&re.RunningInstances) < atomic.LoadInt32(&re.MaxInstances) {
		atomic.AddInt32(&re.RunningInstances, 1)
		return true
	}
	return false
}

//ReleaseToken - release instance token
func (re *RegistryEntry) ReleaseToken() {
	atomic.AddInt32(&re.RunningInstances, -1)
}
