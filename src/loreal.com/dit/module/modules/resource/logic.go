package resource

import (
	"log"
	"os"

	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (m *Module) prepareUploadFolder(uid string) string {
	path := m.UploadPath + "/" + uid + "/"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModeDir|0770); err != nil {
				log.Println("[ERR][Mkdir]", err)
			}
		}
	}
	return path
}

func (m *Module) append(res *Resource) error {
	session, connectionErr := m.MgoSessionManager.Get()
	if connectionErr != nil {
		return connectionErr
	}
	defer session.Close()
	coll := session.DB(m.MgoDbName).C("resource")
	err := coll.Insert(res)
	if err != nil {
		log.Println("[ERR][append]", err)
		return err
	}
	return nil
}

func (m *Module) newRID() bson.ObjectId {
	return bson.NewObjectId()
}

func (m *Module) get(rid string) (*Resource, error) {
	if !bson.IsObjectIdHex(rid) {
		return nil, mgo.ErrNotFound
	}
	session, connectionErr := m.MgoSessionManager.Get()
	if connectionErr != nil {
		return nil, connectionErr
	}
	defer session.Close()
	coll := session.DB(m.MgoDbName).C("resource")

	oid := bson.ObjectIdHex(rid)

	query := coll.FindId(oid)
	res := &Resource{}
	err := query.One(res)
	if err != nil {
		log.Println("[ERR][resource-get]", err)
	}
	return res, err
}

func (m *Module) prepareIndexes() error {
	session, connectionErr := m.MgoSessionManager.Get()
	if connectionErr != nil {
		return connectionErr
	}
	defer session.Close()
	db := session.DB(m.MgoDbName)
	coll := db.C("resource")
	index := mgo.Index{
		Key:        []string{"expires"},
		Background: true,
		Sparse:     true,
	}
	return coll.EnsureIndex(index)
}

func (m *Module) removeExpires() error {
	session, connectionErr := m.MgoSessionManager.Get()
	if connectionErr != nil {
		return connectionErr
	}
	defer session.Close()
	db := session.DB(m.MgoDbName)
	coll := db.C("resource")
	query := bson.M{
		"expires": bson.M{
			"$ne":  time.Unix(0, 0),
			"$lte": time.Now(),
		},
	}
	iter := coll.Find(query).Iter()
	r := Resource{}
	for iter.Next(&r) {
		target := r.FullPath(m.UploadPath, "")
		log.Println("[Cleanup expired files]: Delete", target)
		if err := os.RemoveAll(target); err != nil {
			log.Println("[ERR]: Delete", target, err)
			return err
		}
	}
	if err := iter.Close(); err != nil {
		return err
	}
	if _, err := coll.RemoveAll(query); err != nil {
		log.Println("[ERR]: Remove expired resources from MongoDb", err)
		return err
	}
	return nil
}
