package restful

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

//SQLiteAdapter - SQLite Restful Adapter
type SQLiteAdapter struct {
	DB        *sql.DB
	Mutex     *sync.RWMutex
	TableName string
	tags      []FieldTag
	sqls      map[string]string
	stmts     map[string]*sql.Stmt
	sample    interface{}
}

//NewSQLiteAdapter - create new instance from a model template
func NewSQLiteAdapter(db *sql.DB, mutex *sync.RWMutex, tableName string, modelTemplate interface{}) *SQLiteAdapter {
	if db == nil {
		log.Fatal("[ERR] - [NewSQLiteAdapter] nil db")
	}
	if mutex == nil {
		mutex = &sync.RWMutex{}
	}
	adapter := &SQLiteAdapter{
		DB:        db,
		Mutex:     mutex,
		TableName: tableName,
		tags:      ParseTags(modelTemplate),
		sqls:      make(map[string]string),
		stmts:     make(map[string]*sql.Stmt),
		sample:    modelTemplate,
	}
	if len(adapter.tags) == 0 {
		log.Fatalln("Invalid ModelTemplate:", modelTemplate)
	}
	adapter.init()
	return adapter
}

func (a *SQLiteAdapter) prepareStmt(key, sql string) {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var err error
	if a.stmts[key], err = a.DB.Prepare(sql); err != nil {
		log.Fatal(err)
	}
}

func (a *SQLiteAdapter) init() {
	if _, err := a.DB.Exec(a.createTableSQL()); err != nil {
		log.Printf("[ERR] - [SQLiteAdapter] Can not create table: [%s], err: %v\n", a.TableName, err)
		log.Println("[ERR] - [INFO]", a.createTableSQL())
	}
	createIdxSqls := a.createIndexSQLs()
	for _, cmd := range createIdxSqls {
		_, err := a.DB.Exec(cmd)
		if err != nil {
			log.Printf("[ERR] - [CreateIndex] %v: %s\n", err, cmd)
			return
		}
	}

	a.sqls["set"] = a.setSQL(a.getFields(false /*Do not include ID*/))
	a.sqls["delete"] = a.deleteSQL()
	a.sqls["one"] = a.selectOneSQL()
	for key, sql := range a.sqls {
		if DEBUG {
			log.Printf("[DEBUG] - Prepare [%s]:\n", key)
			fmt.Println("------")
			fmt.Println(sql)
			fmt.Println("------")
			fmt.Println()
		}
		a.prepareStmt(key, sql)
	}
	log.Printf("[INFO] - Table [%s] prepared\n", a.TableName)
}
