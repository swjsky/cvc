package main

import (
	"loreal.com/dit/cmd/logger/modle"
	"time"
)

// Create create logger
func (a *App) Create(l *modle.Logger) (err error) {
	session, err := a.MgoSessionManager.Get()
	if err != nil {
		return
	}
	defer session.Close()
	coll := session.DB(a.Config.MongoDBName).C(loggerCollName)
	l.CreatedAt = time.Now()
	return coll.Insert(l)
}
