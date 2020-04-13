package utils

import (
	"log"
	"sync"

	"gopkg.in/mgo.v2"
)

//MongoSessionManager - manager sessions for mgo
type MongoSessionManager struct {
	ModuleName       string
	MongoConnStr     string
	SecondConnStr    string
	MainSession      *mgo.Session
	SecondSession    *mgo.Session
	SessionPoolLimit int
	mutex            *sync.Mutex
}

//NewMongoSessionManager - create a new MongoSessionManager
func NewMongoSessionManager(moduleName, connectionStr string, poolLimit int) *MongoSessionManager {
	return &MongoSessionManager{
		ModuleName:       moduleName,
		MongoConnStr:     connectionStr,
		SessionPoolLimit: poolLimit,
		mutex:            &sync.Mutex{},
	}
}

//Get clone a mongo session from main session
func (m *MongoSessionManager) Get() (*mgo.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.MainSession == nil {
		var err error
		m.MainSession, err = mgo.Dial(m.MongoConnStr)
		if err != nil {
			return nil, err
		}
		log.Printf("[INFO] - Module:[%s] -> Mongo Session Started.\r\n", m.ModuleName)
		m.MainSession.SetPoolLimit(m.SessionPoolLimit)
	}
	return m.MainSession.Clone(), nil
}

//GetSecond clone a mongo session from 2nd session
func (m *MongoSessionManager) GetSecond(connStr string) (*mgo.Session, error) {
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
		log.Printf("[INFO] - Module:[%s] -> 2nd Mongo Session Started.\r\n", m.ModuleName)
		m.SecondSession.SetPoolLimit(m.SessionPoolLimit)
	}
	return m.SecondSession.Clone(), nil
}

//Dispose - called when disposing the MongoSessionManager
func (m *MongoSessionManager) Dispose() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	log.Println("[INFO] - Disposing Mongo Session Manager.")
	if m.MainSession != nil {
		if err := m.MainSession.Fsync(true); err != nil {
			log.Printf("[ERR] - Module:[%s] -> 1st Mongo session fsync: %v\r\n", m.ModuleName, err)
		}
		log.Printf("[INFO] - Module:[%s] -> Closing 1st Mongo session...\r\n", m.ModuleName)
		m.MainSession.Close()
		log.Printf("[INFO] - Module:[%s] -> 1st Mongo session closed.\r\n", m.ModuleName)
		m.MainSession = nil
	}
	if m.SecondSession != nil {
		if err := m.SecondSession.Fsync(true); err != nil {
			log.Printf("[ERR] - Module:[%s] -> 2nd Mongo session fsync: %v\r\n", m.ModuleName, err)
		}
		log.Printf("[INFO] - Module:[%s] -> Closing 2nd Mongo session...\r\n", m.ModuleName)
		m.SecondSession.Close()
		log.Printf("[INFO] - Module:[%s] -> 2nd Mongo session closed.\r\n", m.ModuleName)
		m.SecondSession = nil
	}
}
