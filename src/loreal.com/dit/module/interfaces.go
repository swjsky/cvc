package module

//Initializable can call to initialize the micro service module
type Initializable interface {
	//Initialize micro service or module
	Init()
}

//Startable can start servce module
type Startable interface {
	//Start micro service or module
	Start()
}

//Stoppable can stop service module
type Stoppable interface {
	//Stop micro service or module
	Stop()
}

//Disposable can dispose resource when exit
type Disposable interface {
	//Dispose resources for micro service or module
	Dispose()
}

//HasRoot can set root
type HasRoot interface {
	SetRoot(*Module)
}

//HasParent has parent/childrean relationship
type HasParent interface {
	SetParent(*Module)
}

//HasSubModule can have sub-modules
type HasSubModule interface {
	Install(subModules ...*Module)
}

//HasStatus can set/get status
type HasStatus interface {
	GetStatus() int32
	SetStatus(int32)
}

//Mountable can mount/unmount resource
type Mountable interface {
	Mount()
	Unmount()
}

//MessageReceiver can receive message
type MessageReceiver interface {
	ReceiveMessage(*Message)
}

//IModule interface for micro service module
type IModule interface {
	HasStatus
	Initializable
	HasRoot
	HasParent
	HasSubModule
	Startable
	Stoppable
	Mountable
	Disposable
	MessageReceiver
}

//ISubModule interface for micro service sub-modules
type ISubModule interface {
	GetBase() *Module
}
