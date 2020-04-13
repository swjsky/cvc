package main

import (
	"fmt"
	"time"

	"github.com/tealeg/xlsx"
)

/*
   openid TEXT DEFAULT '',
    pageid TEXT DEFAULT '',
    scene TEXT DEFAULT '',
    state TEXT INTEGER DEFAULT 0,
    pv INTEGER DEFAULT 0,
    createat DATETIME,
    recent DATETIME
*/

type visitRecord struct {
	OpenID   string
	PageID   string
	Scene    string
	PV       int
	CreateAt time.Time
	Recent   time.Time
}

func (a *App) genVisitReportXlsx(
	runtime *RuntimeEnv,
	from time.Time,
	to time.Time,
) (f *xlsx.File, err error) {
	const stmtName = "visit-report"
	const stmtSQL = "SELECT openid,pageid,scene,pv,createat,recent FROM visit WHERE createat>=? AND createat<=?;"
	stmt := a.getStmt(runtime, stmtName)
	if stmt == nil {
		//lazy setup for stmt
		if stmt, err = a.setStmt(runtime, stmtName, stmtSQL); err != nil {
			return
		}
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	rows, err := stmt.Query(from, to)
	if err != nil {
		return
	}

	var headerRow *xlsx.Row

	f = xlsx.NewFile()
	sheet, err := f.AddSheet(fmt.Sprintf("PV details %s-%s", from.Format("0102"), to.Format("0102")))
	if err != nil {
		return nil, err
	}
	headerRow = sheet.AddRow()
	headerRow.AddCell().SetString("OpenID")
	headerRow.AddCell().SetString("页面号")
	headerRow.AddCell().SetString("入口场景")
	headerRow.AddCell().SetString("PV")
	headerRow.AddCell().SetString("首次进入时间")
	headerRow.AddCell().SetString("最后进入时间")

	for rows.Next() {
		r := sheet.AddRow()
		vr := visitRecord{}
		if err = rows.Scan(
			&vr.OpenID,
			&vr.PageID,
			&vr.Scene,
			&vr.PV,
			&vr.CreateAt,
			&vr.Recent,
		); err != nil {
			return
		}
		r.AddCell().SetString(vr.OpenID)
		r.AddCell().SetString(vr.PageID)
		r.AddCell().SetString(vr.Scene)
		r.AddCell().SetInt(vr.PV)
		r.AddCell().SetDateTime(vr.CreateAt)
		r.AddCell().SetDateTime(vr.Recent)
	}
	return f, nil
}
