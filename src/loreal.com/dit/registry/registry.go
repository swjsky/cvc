package registry

//Registry - interface for any registry
type Registry interface {
	Register(Name, Description, Addr string, Tags ...string) (unRegisterSignal chan bool)
	Unregister(unRegisterSignal chan bool)
}
