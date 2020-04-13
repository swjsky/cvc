package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

//InitDB - initialized database
func (a *App) InitDB() {
	//init database tables
	sqlStmts := []string{
		//PV计数
		`CREATE TABLE IF NOT EXISTS visit (
    openid TEXT DEFAULT '',
    pageid TEXT DEFAULT '',
    scene TEXT DEFAULT '',
    state TEXT INTEGER DEFAULT 0,
    pv INTEGER DEFAULT 0,
    createat DATETIME,
    recent DATETIME
);`,
		"CREATE INDEX IF NOT EXISTS idx_visit_openid ON visit(openid);",
		"CREATE INDEX IF NOT EXISTS idx_visit_pageid ON visit(pageid);",
		"CREATE INDEX IF NOT EXISTS idx_visit_scene ON visit(scene);",
		"CREATE INDEX IF NOT EXISTS idx_visit_state ON visit(state);",
		"CREATE INDEX IF NOT EXISTS idx_visit_createat ON visit(createat);",
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
