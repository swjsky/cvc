//Package task - task manager for backend long run program logic (go routings)
package task

//State - running state for task instance
type State int

const (
	//StateNew - State New
	StateNew State = iota
	//StateRunning - State Running
	StateRunning
	//StatePaused - State Paused
	StatePaused
	//StateCompleted - State Completed
	StateCompleted
	//StateCanceled - State Canceled
	StateCanceled
)

func (s *State) String() string {
	switch *s {
	case StateNew:
		return "new"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateCompleted:
		return "completed"
	case StateCanceled:
		return "canceled"
	}
	return ""
}
