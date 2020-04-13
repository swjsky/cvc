package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

//reportVisitHandler - export visit detail report in excel format
//endpoint: /report/visit
//method: GET
func (a *App) reportVisitHandler(w http.ResponseWriter, r *http.Request) {
	const dateFormat = "2006-01-02"
	if r.Method != "GET" {
		showError(w, r, "明细报表下载", "调用方法不正确")
		return
	}
	q := r.URL.Query()
	env := a.getEnv(q.Get("appid"))
	rt := a.getRuntime(env)
	if rt == nil {
		showError(w, r, "明细报表下载", "参数错误， APPID不正确")
		return
	}
	var err error
	from, err := time.ParseInLocation(dateFormat, sanitizePolicy.Sanitize(q.Get("from")), time.Local)
	if err != nil {
		showError(w, r, "明细报表下载", "参数错误， 开始时间‘from’格式不正确")
		return
	}
	to, err := time.ParseInLocation(dateFormat, sanitizePolicy.Sanitize(q.Get("to")), time.Local)
	if err != nil {
		showError(w, r, "明细报表下载", "参数错误， 结束时间‘to’格式不正确")
		return
	}
	to = to.Add(time.Second*(60*60*24-1) + time.Millisecond*999) //23:59:59.999
	xlsxFile, err := a.genVisitReportXlsx(rt, from, to)
	if err != nil {
		log.Println("[ERR] - [ep][report/download], err:", err)
		showError(w, r, "明细报表下载", "查询报表时发生错误， 请联系管理员查看日志。")
		return
	}
	fileName := fmt.Sprintf("visit-report-%s-%s.xlsx", from.Format("0102"), to.Format("0102"))
	w.Header().Add("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Add("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := xlsxFile.Write(w); err != nil {
		showError(w, r, "明细报表下载", "查询报表时发生错误， 无法生存Xlsx文件。")
		return
	}
}
