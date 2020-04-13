package websocket

import (
	"loreal.com/dit/utils"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ActionType action type
type ActionType int

const (
	// CONNSTATUSOPEN websocket connection open
	CONNSTATUSOPEN = iota
	// CONNSTATUSCLOSED websocket connection closed
	CONNSTATUSCLOSED

	// MAXMESSAGESIZE MAXMESSAGESIZE
	MAXMESSAGESIZE = 8192
	// WRITETIMEOUT write time out
	WRITETIMEOUT = 10 * time.Second
	// PONGWAIT PONGWAIT
	PONGWAIT = 12 * time.Second
	// PINGPERIOD PINGPERIOD
	PINGPERIOD = (12 * time.Second * 9) / 10
)

const (
	// ActionTypeJoin join websocket manager
	ActionTypeJoin ActionType = iota
	// ActionTypeLeave leave websocket manager
	ActionTypeLeave
	// ActionTypeReceive receive websocket manager
	ActionTypeReceive
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// CheckOriginURL check origin
func CheckOriginURL(r *http.Request) bool {
	return true
}

// WsAction single websocket action
type WsAction struct {
	ActionType ActionType
	Websocket  *Websocket
	Message    string
}

var actionChan = make(chan *WsAction, 100)

// Notification websocket Notification
type Notification interface {
	AccountID() string
	Contents() []byte
}

/////////////////////////////////////////////////////
// WebsocketManager

// Manager websocket manager
type Manager struct {
	Websockets map[string][]*Websocket
	Count      int
	ToSend     chan Notification
	ToReceive  chan string
	ToSkip     chan bool
	sync       *sync.Mutex
	quit       chan struct{}
}

// NewWebsocketManager return websocket manager
func NewWebsocketManager() *Manager {
	return &Manager{
		Websockets: make(map[string][]*Websocket),
		ToSend:     make(chan Notification, 100),
		ToReceive:  make(chan string, 100),
		ToSkip:     make(chan bool, 100),
		sync:       &sync.Mutex{},
		quit:       make(chan struct{}, 1),
	}
}

// Listen websocket manager listern
func (m *Manager) Listen() {
	for {
		select {
		case message, ok := <-m.ToSend:
			if ok {
				m.HandleToSend(message)
			}
		case actionCenter, ok := <-actionChan:
			if ok {
				switch actionCenter.ActionType {
				case ActionTypeJoin:
					m.Join(actionCenter.Websocket)
				case ActionTypeLeave:
					m.Leave(actionCenter.Websocket)
				case ActionTypeReceive:
					m.ToReceive <- actionCenter.Message
				}
			}
		case <-m.quit:
			m.sync.Lock()
			defer m.sync.Unlock()
			defer close(m.quit)
			defer close(actionChan)
			for _, websockets := range m.Websockets {
				for _, websocket := range websockets {
					websocket.Close()
				}
			}
			return
		}
	}
}

// HandleToSend handler to websocket
func (m *Manager) HandleToSend(message Notification) {
	if len(m.Websockets[message.AccountID()]) == 0 {
		m.ToSkip <- true
		return
	}
	if websockets, ok := m.Websockets[message.AccountID()]; ok {
		wg := sync.WaitGroup{}
		for _, websocket := range websockets {
			wg.Add(1)
			go func(w *Websocket, msg Notification) {
				w.Notification <- msg
				wg.Done()
			}(websocket, message)
		}
		wg.Wait()
	}
}

// Join join websocket to websocket manager
func (m *Manager) Join(w *Websocket) {
	m.sync.Lock()
	defer m.sync.Unlock()

	if websockets, ok := m.Websockets[w.AccountID]; ok {
		websockets = append(websockets, w)
		m.Websockets[w.AccountID] = websockets
	} else {
		m.Websockets[w.AccountID] = []*Websocket{w}
	}
	m.Count++
	log.Printf("New websocket connection from Account %s, Total connections: %d", w.AccountID, len(m.Websockets))
}

// Leave remove websocket from websocket manager
func (m *Manager) Leave(w *Websocket) {
	m.sync.Lock()
	defer m.sync.Unlock()

	if websockets, ok := m.Websockets[w.AccountID]; ok {
		_websockets := []*Websocket{}
		for _, connected := range websockets {
			if connected.ID != w.ID {
				_websockets = append(_websockets, connected)
			}
		}
		if len(_websockets) == 0 {
			delete(m.Websockets, w.AccountID)
		} else {
			m.Websockets[w.AccountID] = _websockets
		}
		m.Count--
		log.Printf("Websocket %s leave, Total connections: %d", w.AccountID, len(m.Websockets))
	} else {
		log.Printf("Websocket connection %s is not exists", w.AccountID)
	}
}

// Close close websocket conn
func (m *Manager) Close() {
	m.quit <- struct{}{}
}

/////////////////////////////////////////////////////
// Websocket

// Websocket websocket
type Websocket struct {
	ID           string
	AccountID    string
	Connection   *websocket.Conn
	Notification chan Notification
	QuitWrite    chan bool
	QuitRead     chan bool
	sync         *sync.Mutex
	status       uint8
}

// NewWebsocket return websocket
func NewWebsocket(accountID string, conn *websocket.Conn) *Websocket {
	websocket := &Websocket{
		ID:           utils.RandomString(50),
		AccountID:    accountID,
		Connection:   conn,
		Notification: make(chan Notification, 1),
		QuitWrite:    make(chan bool, 1),
		QuitRead:     make(chan bool, 1),
		sync:         new(sync.Mutex),
		status:       CONNSTATUSOPEN,
	}

	actionChan <- &WsAction{
		ActionType: ActionTypeJoin,
		Websocket:  websocket,
	}
	return websocket
}

// Listen listen websocket，start with endpoint create
func (w *Websocket) Listen() {
	go w.writePump()

	w.readPump()
}

func (w *Websocket) readPump() {
	w.Connection.SetReadLimit(MAXMESSAGESIZE)
	w.Connection.SetReadDeadline(time.Now().Add(PONGWAIT))
	w.Connection.SetCloseHandler(nil)
	w.Connection.SetPongHandler(func(appdata string) error {
		w.Connection.SetReadDeadline(time.Now().Add(PONGWAIT))
		return nil
	})

	go func(ws *Websocket) {
		defer ws.Close()
		for {
			_, message, err := ws.Connection.ReadMessage()
			if err != nil {
				log.Println(err.Error())
				break
			}
			actionChan <- &WsAction{
				ActionType: ActionTypeReceive,
				Message:    string(message),
			}
		}
	}(w)
	<-w.QuitRead
}

func (w *Websocket) writePump() {
	ticker := time.NewTicker(PINGPERIOD)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case notification, ok := <-w.Notification:
			if ok {
				err := w.write(websocket.TextMessage, notification.Contents())
				if err != nil {
					log.Printf("Write websocket of %s get an error: %s", w.AccountID, err.Error())
				}
			}
		case <-ticker.C:
			if err := w.write(websocket.PingMessage, []byte{}); err != nil {
				log.Println(err.Error())
				w.Close()
			}
		// 跳出循环
		case <-w.QuitWrite:
			return
		}
	}
}

// Close close websocket
func (w *Websocket) Close() {
	w.sync.Lock()
	defer w.sync.Unlock()
	if w.status == CONNSTATUSOPEN {
		close(w.Notification)
		w.Connection.Close()
		w.status = CONNSTATUSCLOSED
		actionChan <- &WsAction{
			ActionType: ActionTypeLeave,
			Websocket:  w,
		}
		w.QuitWrite <- true
		w.QuitRead <- true
		close(w.QuitWrite)
		close(w.QuitRead)
		log.Printf("Websocket %s close", w.AccountID)
	}
}

func (w *Websocket) write(length int, playload []byte) (err error) {
	w.Connection.SetWriteDeadline(time.Now().Add(WRITETIMEOUT))
	return w.Connection.WriteMessage(length, playload)
}
