package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gobuffalo/packr/v2"
	_ "github.com/mattn/go-sqlite3"
	migrate "github.com/rubenv/sql-migrate"
)

//InitDB - initialized database
func (a *App) InitDB() {
	//init database tables

	var err error
	for _, env := range a.Runtime {
		env.db, err = sql.Open("sqlite3", fmt.Sprintf("%s%s?cache=shared&mode=rwc", env.Config.DataFolder, env.Config.SqliteDB))
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("[INFO] - applying database  [%s] migrations...\n", env.Config.Name)

		// migrations := &migrate.FileMigrationSource{
		// 	Dir: "sql-migrations",
		// }
		migrations := &migrate.PackrMigrationSource{
			Box: packr.New("sql-migrations", "./sql-migrations"),
		}
		n, err := migrate.Exec(env.db, "sqlite3", migrations, migrate.Up)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("[INFO] - [%d] migration file applied\n", n)

		log.Printf("[INFO] - DB for [%s] ready!\n", env.Config.Name)
	}
}
