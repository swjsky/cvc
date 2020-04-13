/*Package sqlitehelper - helper functions for sqlite db
  includes:
  BatchRunner - to improve performance
*/
package sqlitehelper

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	//sqlite3
	_ "github.com/mattn/go-sqlite3"
)

//SQLStmtInstance - instance of one sql stmt
type SQLStmtInstance struct {
	Stmt *sql.Stmt
	Args []interface{}
}

//SQLErrorInstance - instance of one sql error
type SQLErrorInstance struct {
	Name    string
	Args    []interface{}
	ErrType string
	Err     error
}

//BatchRunner - run sqlite query in batch
/*
   //Example
   func batchInsert(tagID string, openIDs []string) error {
       const insertSQL = "INSERT OR REPLACE INTO tagdata_%d (tagid,openid) VALUES (?,?);"

       sqlRunner := NewBatchRunner(t.Owner.db, 10000, t.Owner.mutex)
       if err := sqlRunner.prepare("cleanup", fmt.Sprintf(cleanupSQL, tagID)); err != nil {
           return err
       }
       if err := sqlRunner.prepare("insert", fmt.Sprintf(insertSQL, tagID)); err != nil {
           return err
       }
       //Start 3 writer
       sqlRunner.Start(3)
       defer func() {
           //Close will commit any SQL stmts in buffer before close.
           sqlRunner.Close()
       }()

       //Run cleanup
       sqlRunner.Push("cleanup", tagID)
       sqlRunner.Commit()

       for _, openid := range openIDs {
           sqlRunner.Push("insert", tagID, openid)
       }
       sqlRunner.Commit()
   }
*/
type BatchRunner struct {
	DB          *sql.DB
	BatchSize   int32
	Stmts       map[string]*sql.Stmt
	Names       map[*sql.Stmt]string
	CommandChan chan struct {
		Name string
		Args []interface{}
	}
	ErrChan     chan SQLErrorInstance
	ErrHandlers []func(SQLErrorInstance)
	Runners     map[uint32]chan string
	RunnerCount uint32
	ErrorCount  uint32
	MaxErrors   int
	DBMutex     *sync.RWMutex
	Mutex       *sync.RWMutex
	wg          *sync.WaitGroup
}

//NewBatchRunner - constructor
func NewBatchRunner(db *sql.DB, batchSize int32, dbMutex *sync.RWMutex) *BatchRunner {
	const defaultErrorBuffer = 10
	const defaultMaxErrors = -1
	return &BatchRunner{
		DB:        db,
		BatchSize: batchSize,
		Stmts:     make(map[string]*sql.Stmt, 2),
		Names:     make(map[*sql.Stmt]string, 2),
		CommandChan: make(chan struct {
			Name string
			Args []interface{}
		}, 0),
		ErrChan:     make(chan SQLErrorInstance, defaultErrorBuffer),
		ErrHandlers: make([]func(SQLErrorInstance), 0, 1),
		Runners:     make(map[uint32]chan string),
		MaxErrors:   defaultMaxErrors,
		DBMutex:     dbMutex,
		Mutex:       &sync.RWMutex{},
		wg:          &sync.WaitGroup{},
	}
}

//prepare - prepare named query into runner
func (r *BatchRunner) Prepare(name, query string, args ...interface{}) (err error) {
	//prepare in DB
	var stmt *sql.Stmt
	r.DBMutex.Lock()
	stmt, err = r.DB.Prepare(fmt.Sprintf(query, args...))
	if err != nil {
		r.DBMutex.Unlock()
		log.Println("[ERR] - [SQLiteRunner] prepare", err)
		r.reportError(name, "prepare", err, query)
		return
	}
	r.DBMutex.Unlock()
	//store into map
	r.Mutex.Lock()
	r.Stmts[name] = stmt
	r.Names[stmt] = name
	r.Mutex.Unlock()
	return
}

//RegisterErrHandler - register error handler
func (r *BatchRunner) RegisterErrHandler(handler func(SQLErrorInstance)) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.ErrHandlers = append(r.ErrHandlers, handler)
}

//Push - push query instance into runner
func (r *BatchRunner) Push(name string, args ...interface{}) {
	r.CommandChan <- struct {
		Name string
		Args []interface{}
	}{
		Name: name,
		Args: args,
	}
}

//Query - run query on underline DB
func (r *BatchRunner) Query(name string, args ...interface{}) (*sql.Rows, error) {
	r.Mutex.RLock()
	stmt, ok := r.Stmts[name]
	r.Mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("[ERR] - [SQLiteRunner] Missing Stmt: %v", name)
	}
	r.DBMutex.Lock()
	defer r.DBMutex.Unlock()
	return stmt.Query(args...)
}

//QueryRow - run QueryRow on underline DB
func (r *BatchRunner) QueryRow(name string, args ...interface{}) (*sql.Row, error) {
	r.Mutex.RLock()
	stmt, ok := r.Stmts[name]
	r.Mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("[ERR] - [SQLiteRunner] Missing Stmt: %v", name)
	}
	r.DBMutex.Lock()
	defer r.DBMutex.Unlock()
	return stmt.QueryRow(args...), nil
}

//reportError - record err and return whether to continue
func (r *BatchRunner) reportError(name, errType string, err error, args ...interface{}) (Continue bool) {
	errorCount := int(atomic.AddUint32(&r.ErrorCount, 1))
	r.ErrChan <- SQLErrorInstance{
		Name:    name,
		Args:    args,
		ErrType: errType,
		Err:     err,
	}
	return r.MaxErrors == -1 || errorCount <= r.MaxErrors
}

//Commit - send $commit command to runner on ctrlChan
func (r *BatchRunner) Commit() {
	r.Mutex.RLock()
	for _, ctrlBus := range r.Runners {
		ctrlBus <- "$commit"
		// log.Println("[INFO] - [SQLiteRunner] Sent $commit to runner:", runnerID)
	}
	r.Mutex.RUnlock()
}

//doCommit - do commit on any querys in the buffer
func (r *BatchRunner) doCommit(buffer []SQLStmtInstance) error {
	//Start a new transaction
	// log.Println("$commit", len(buffer), cap(buffer))
	r.DBMutex.Lock()
	tx, err := r.DB.Begin()
	if err != nil {
		r.reportError("tx.begin", "tx.begin", err)
		log.Println("[ERR] - [SQLiteRunner]Begin transaction", err)
		return err
	}
	defer func(tx *sql.Tx, mutex *sync.RWMutex) {
		if err := tx.Commit(); err != nil {
			r.reportError("tx.commit", "tx.commit", err)
			log.Println("[ERR] - [SQLiteRunner]Commit transaction", err)
		}
		mutex.Unlock()
	}(tx, r.DBMutex)
	for _, stmt := range buffer {
		if _, err := tx.Stmt(stmt.Stmt).Exec(stmt.Args...); err != nil {
			r.Mutex.RLock()
			name := r.Names[stmt.Stmt]
			r.Mutex.RUnlock()
			r.reportError(name, "stmt", err, stmt.Args...)
			log.Println("[ERR] - [SQLiteRunner] Running stmt", err)
		}
	}
	return nil
}

//Close - close backend writers
func (r *BatchRunner) Close() {
	close(r.CommandChan)
	r.wg.Wait()
	close(r.ErrChan)
	r.Mutex.Lock()
	for name, v := range r.Stmts {
		if err := v.Close(); err != nil {
			log.Println("[ERR] - Close stms", name, err)
		}
	}
	r.Mutex.Unlock()
	log.Println("[INFO] - [SQLiteRunner] Closed")
}

//Start - total 'numberOfInstances' backend writers
func (r *BatchRunner) Start(numberOfInstances int) {
	if numberOfInstances <= 0 {
		numberOfInstances = 1
	}
	for cnt := 0; cnt < numberOfInstances; cnt++ {
		r.doStart()
	}
	//Start errHandler
	go r.errHandler()
}

//errHandler - go runting to handle errors
func (r *BatchRunner) errHandler() {
	for e := range r.ErrChan {
		r.Mutex.RLock()
		handlers := r.ErrHandlers
		r.Mutex.RUnlock()
		for _, h := range handlers {
			if h != nil {
				h(e)
			}
		}
	}
}

//doStart - start one instance of backend writer
func (r *BatchRunner) doStart() {
	runnerID := atomic.LoadUint32(&r.RunnerCount)
	atomic.AddUint32(&r.RunnerCount, 1)
	ctrlChan := make(chan string, 0)
	r.Mutex.Lock()
	r.Runners[runnerID] = ctrlChan
	r.Mutex.Unlock()
	log.Printf("[INFO] - start SQLite writer: [%d]\r\n", runnerID)
	go func(r *BatchRunner, ctrlChan chan string) {
		r.wg.Add(1)
		defer r.wg.Done()
		buffer := make([]SQLStmtInstance, 0, r.BatchSize)
		for {
			select {
			case ctrlCmd := <-ctrlChan:
				switch ctrlCmd {
				case "$commit":
					r.doCommit(buffer[:len(buffer)])
					buffer = buffer[0:0]
				}
			case cmd, ok := <-r.CommandChan:
				if !ok {
					goto exit
				}
				r.Mutex.RLock()
				stmt, stmtOk := r.Stmts[cmd.Name]
				batchSize := int(r.BatchSize)
				r.Mutex.RUnlock()
				if !stmtOk {
					log.Println("[ERR] - [SQLiteRunner]missing query:", cmd.Name)
					continue
				}
				buffer = append(buffer, SQLStmtInstance{
					Stmt: stmt,
					Args: cmd.Args,
				})
				if len(buffer) >= batchSize {
					r.doCommit(buffer[:len(buffer)])
					buffer = buffer[0:0]
				}

			}
		}
	exit:
		if len(buffer) > 0 {
			r.doCommit(buffer[:len(buffer)])
		}
		buffer = buffer[0:0]
	}(r, ctrlChan)
}
