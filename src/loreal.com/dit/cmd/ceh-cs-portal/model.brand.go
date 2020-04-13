package main

import "time"

//Brand - Loreal Brand
type Brand struct {
	ID         int64     `name:"id" type:"INTEGER"`
	Code       string    `type:"TEXT" index:"asc"`
	Name       string    `type:"TEXT" index:"asc"`
	CreateAt   time.Time `type:"DATETIME" default:"datetime('now','localtime')"`
	Modified   time.Time `type:"DATETIME"`
	CreateBy   string    `type:"TEXT"`
	ModifiedBy string    `type:"TEXT"`
}
