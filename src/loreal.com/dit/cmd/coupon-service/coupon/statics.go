package coupon

import (
	"database/sql"
)

var dbConnection      				*sql.DB

func staticsInit(databaseConnection  	*sql.DB) {
	dbConnection = databaseConnection
}