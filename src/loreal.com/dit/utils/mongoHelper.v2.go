package utils

import (
	"log"
	"sync"

	"github.com/globalsign/mgo"
)

//MongoSessionManagerV2 - manager sessions for mgo
type MongoSessionManagerV2 struct {
	ModuleName       string
	MongoConnStr     string
	SecondConnStr    string
	MainSession      *mgo.Session
	SecondSession    *mgo.Session
	SessionPoolLimit int
	mutex            *sync.Mutex
}

//NewMongoSessionManagerV2 - create a new MongoSessionManagerV2
func NewMongoSessionManagerV2(moduleName, connectionStr string, poolLimit int) *MongoSessionManagerV2 {
	return &MongoSessionManagerV2{
		ModuleName:       moduleName,
		MongoConnStr:     connectionStr,
		SessionPoolLimit: poolLimit,
		mutex:            &sync.Mutex{},
	}
}

//Get clone a mongo session from main session
func (m *MongoSessionManagerV2) Get() (*mgo.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.MainSession == nil {
		var err error
		m.MainSession, err = mgo.Dial(m.MongoConnStr)
		if err != nil {
			return nil, err
		}
		log.Printf("[INFO] - Module:[%s] -> Mongo Session-V2 Started.\r\n", m.ModuleName)
		m.MainSession.SetPoolLimit(m.SessionPoolLimit)
	}
	return m.MainSession.Clone(), nil
}

//GetSecond clone a mongo session from 2nd session
func (m *MongoSessionManagerV2) GetSecond(connStr string) (*mgo.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.SecondConnStr != connStr {
		m.SecondConnStr = connStr
		if m.SecondSession != nil {
			m.SecondSession.Close()
			m.SecondSession = nil
		}
	}
	if m.SecondSession == nil {
		var err error
		m.SecondSession, err = mgo.Dial(m.SecondConnStr)
		if err != nil {
			return nil, err
		}
		log.Printf("[INFO] - Module:[%s] -> 2nd Mongo Session-V2 Started.\r\n", m.ModuleName)
		m.SecondSession.SetPoolLimit(m.SessionPoolLimit)
	}
	return m.SecondSession.Clone(), nil
}

//Dispose - called when disposing the MongoSessionManagerV2
func (m *MongoSessionManagerV2) Dispose() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	log.Println("[INFO] - Disposing Mongo Session Manager V2.")
	if m.MainSession != nil {
		if err := m.MainSession.Fsync(true); err != nil {
			log.Printf("[ERR] - Module:[%s] -> 1st Mongo session-V2 fsync: %v\r\n", m.ModuleName, err)
		}
		log.Printf("[INFO] - Module:[%s] -> Closing 1st Mongo session-V2...\r\n", m.ModuleName)
		m.MainSession.Close()
		log.Printf("[INFO] - Module:[%s] -> 1st Mongo session-V2 closed.\r\n", m.ModuleName)
		m.MainSession = nil
	}
	if m.SecondSession != nil {
		if err := m.SecondSession.Fsync(true); err != nil {
			log.Printf("[ERR] - Module:[%s] -> 2nd Mongo session-V2 fsync: %v\r\n", m.ModuleName, err)
		}
		log.Printf("[INFO] - Module:[%s] -> Closing 2nd Mongo session-V2...\r\n", m.ModuleName)
		m.SecondSession.Close()
		log.Printf("[INFO] - Module:[%s] -> 2nd Mongo session-V2 closed.\r\n", m.ModuleName)
		m.SecondSession = nil
	}
}
