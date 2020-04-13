package module

import "time"

//MessageHandler define handler function for message
type MessageHandler func(msgPtr *Message) (handled bool)

//MessageCallback callback func for message
type MessageCallback func(msgPtr *Message)

//Message - Micro service message for CEH
type Message struct {
	Source     string
	Name       string
	Args       []interface{}
	Callback   MessageCallback
	ResultChan chan interface{}
	Results    []interface{}
	Err        error
	Broadcast  bool
}

//NewMessage create new message
func NewMessage(source string, name string, callback MessageCallback, args ...interface{}) *Message {
	return &Message{
		Source:   source,
		Name:     name,
		Args:     args,
		Callback: callback,
	}
}

//AddMessageHandler add message handler by message name
func (p *Module) AddMessageHandler(messageName string, h MessageHandler) {
	if status := p.GetStatus(); status == StatusStarting || status == StatusRunning || status == StatusStopping {
		//Cannot modify when module is running
		return
	}
	var handlers []MessageHandler
	var ok bool
	if handlers, ok = p.MessageHandlers[messageName]; ok {
		handlers = append(handlers, h)
	} else {
		handlers = []MessageHandler{h}
	}
	p.MessageHandlers[messageName] = handlers
}

func (p *Module) startMessageLoop() {
	ticker := time.NewTicker(p.WatchDogTick)
	for {
		select {
		case msgPtr, ok := <-p.MessageBus:
			if !ok {
				p.SetStatus(StatusStopped)
				break
			}
			p.wg.Add(1)
			go p.processMessage(msgPtr)
		case <-ticker.C:
			//log.Printf("[Watch Dog] %s:\r\n", p.String())
			if p.OnTick != nil {
				go p.OnTick(p)
			}
		}
	}
}

//ReceiveMessage into MessageBus
func (p *Module) ReceiveMessage(msgPtr *Message) {
	p.MessageBus <- msgPtr
}

func (p *Module) processMessage(msgPtr *Message) {
	defer p.wg.Done()
	var handled bool
	if handlers, ok := p.MessageHandlers[msgPtr.Name]; ok {
		//if multi-handlers registered for a message, the message considered to be handled when any one of handler returns true
		for _, handler := range handlers {
			//log.Printf("Handler %d\n", i+1)
			handled = handler(msgPtr) || handled
		}
	}
	if handled && !msgPtr.Broadcast {
		//log.Printf("Handled\n")
		if msgPtr.Callback != nil {
			msgPtr.Callback(msgPtr)
		}
		return
	}
	for _, c := range p.Children {
		if messageReceiver, ok := c.(MessageReceiver); ok {
			//log.Printf("Message [%s] delieved to %s\r\n", msgPtr.Name, c)
			messageReceiver.ReceiveMessage(msgPtr)
		}
	}
}

//Send message to current micro service module with a callback
func (p *Module) Send(name string, callback MessageCallback, args ...interface{}) {
	target := p
	if p.Root != nil {
		target = p.Root
	}
	target.MessageBus <- &Message{
		Source:   target.String(),
		Name:     name,
		Args:     args,
		Callback: callback,
	}
}

//SendWithChan send message to current micro service module with a result chan
func (p *Module) SendWithChan(name string, resultChan chan interface{}, args ...interface{}) {
	target := p
	if p.Root != nil {
		target = p.Root
	}
	target.MessageBus <- &Message{
		Source:     target.String(),
		Name:       name,
		Args:       args,
		ResultChan: resultChan,
		Callback:   nil,
	}
}

//SendTo message to micro service module
func (p *Module) SendTo(target *Module, name string, callback MessageCallback, args ...interface{}) {
	target.MessageBus <- &Message{
		Source:   p.String(),
		Name:     name,
		Args:     args,
		Callback: callback,
	}
}

//Broadcast message
func (p *Module) Broadcast(name string, callback MessageCallback, args ...interface{}) {
	target := p
	if p.Root != nil {
		target = p.Root
	}
	target.MessageBus <- &Message{
		Source:    target.String(),
		Name:      name,
		Args:      args,
		Callback:  callback,
		Broadcast: true,
	}
}
