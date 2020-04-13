package main

import (
	"database/sql"
	"fmt"
	"log"
	// _ "github.com/mattn/go-sqlite3"
)

//InitDB - initialized database
func (a *App) InitDB() {
	//init database tables
	sqlStmts := []string{
		//PV计数
		`CREATE TABLE IF NOT EXISTS Visit (
	UserID INTEGER DEFAULT 0,
	PageID TEXT DEFAULT '',
	Scene TEXT DEFAULT '',
	State TEXT INTEGER DEFAULT 0,
	PV INTEGER DEFAULT 0,
	CreateAt DATETIME,
	Recent DATETIME
);`,
		"CREATE INDEX IF NOT EXISTS idxVisitUserID ON Visit(UserID);",
		"CREATE INDEX IF NOT EXISTS idxVisitPageID ON Visit(PageID);",
		"CREATE INDEX IF NOT EXISTS idxVisitScene ON Visit(Scene);",
		"CREATE INDEX IF NOT EXISTS idxVisitState ON Visit(State);",
		"CREATE INDEX IF NOT EXISTS idxVisitCreateAt ON Visit(CreateAt);",
	}

	var err error
	for _, env := range a.Runtime {
		env.db, err = sql.Open("sqlite3", fmt.Sprintf("%s%s?cache=shared&mode=rwc", env.Config.DataFolder, env.Config.SqliteDB))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("[INFO] - Initialization DB for [%s]...\n", env.Config.Name)
		for _, sqlStmt := range sqlStmts {
			_, err := env.db.Exec(sqlStmt)
			if err != nil {
				log.Printf("[ERR] - [InitDB] %q: %s\n", err, sqlStmt)
				return
			}
		}
		env.stmts = make(map[string]*sql.Stmt, 0)
		log.Printf("[INFO] - DB for [%s] ready!\n", env.Config.Name)
	}
}
