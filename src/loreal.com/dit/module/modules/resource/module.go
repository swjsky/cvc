package resource

import (
	"log"
	"time"

	"loreal.com/dit/middlewares"
	"loreal.com/dit/module"
	"loreal.com/dit/utils"
)

const sessionPoolLimit = 1000
const moduleName = "resource"
const moduleDescription = "CEH resource module"

//Module - Micro service module for wechat JSAPI
type Module struct {
	*module.Module
	MgoSessionManager *utils.MongoSessionManager
	MgoDbName         string
	UploadPath        string
	MimeHandlers      []MimeHandler
}

//GetBase - implements ISubModule interface
func (m *Module) GetBase() *module.Module {
	return m.Module
}

//NewModule create new wechat module
func NewModule(path, mongoConnStr, mongoDbName, UploadPath string, userStore middlewares.RoleVerifier, handlers ...MimeHandler) *Module {
	m := &Module{
		Module:            module.NewModule(moduleName, moduleDescription, path),
		MgoSessionManager: utils.NewMongoSessionManager(moduleName, mongoConnStr, sessionPoolLimit),
		MgoDbName:         mongoDbName,
		UploadPath:        UploadPath,
		MimeHandlers:      handlers,
	}
	m.registerHandlers()
	m.registerEndpoints(userStore)
	m.WatchDogTick = 5 * time.Minute
	m.OnTick = func(p *module.Module) {
		//log.Println("Cleanup expired resources...")
		m.removeExpires()
	}
	if err := m.prepareIndexes(); err != nil {
		log.Println("[ERR]: prepareIndexes()", err)
	}
	return m
}
