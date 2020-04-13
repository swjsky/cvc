package main

func (a *App) recordPV(
	runtime *RuntimeEnv,
	openid, pageid, scene string,
	visitState int,
) (err error) {
	const stmtNameNewPV = "insert-visit"
	const stmtSQLNewPV = "INSERT INTO visit (openid,pageid,scene,createAt,recent,pv,state) VALUES (?,?,?,datetime('now','localtime'),datetime('now','localtime'),1,?);"
	const stmtNamePV = "update-pv"
	const stmtSQLPV = "UPDATE visit SET pv=pv+1,recent=datetime('now','localtime') WHERE openid=? AND pageid=? AND scene=? AND state=?;"
	stmtPV := a.getStmt(runtime, stmtNamePV)
	if stmtPV == nil {
		if stmtPV, err = a.setStmt(runtime, stmtNamePV, stmtSQLPV); err != nil {
			return
		}
	}
	stmtNewPV := a.getStmt(runtime, stmtNameNewPV)
	if stmtNewPV == nil {
		if stmtNewPV, err = a.setStmt(runtime, stmtNameNewPV, stmtSQLNewPV); err != nil {
			return
		}
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	tx, err := runtime.db.Begin()
	if err != nil {
		return err
	}
	stmtPV = tx.Stmt(stmtPV)
	pvResult, err := stmtPV.Exec(
		openid,
		pageid,
		scene,
		visitState,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	cnt, err := pvResult.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if cnt > 0 {
		tx.Commit()
		return
	}
	stmtNewPV = tx.Stmt(stmtNewPV)
	_, err = stmtNewPV.Exec(
		openid,
		pageid,
		scene,
		visitState,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return
}

// func (a *App) recordQRScan(
// 	openid string,
// ) (err error) {
// 	for _, env := range a.Config.Envs {
// 		rt := a.getRuntime(env.Name)
// 		if rt == nil {
// 			continue
// 		}
// 		a.doRecordQRScan(rt, openid)
// 	}
// 	return nil
// }

// func (a *App) doRecordQRScan(
// 	runtime *RuntimeEnv,
// 	openid string,
// ) (err error) {
// 	const stmtNameAdd = "add-openid"
// 	const stmtSQLAdd = "INSERT INTO qrscan (openid,createat) VALUES (?,datetime('now','localtime'));"
// 	const stmtNameRecord = "record-scan"
// 	const stmtSQLRecord = "UPDATE qrscan SET scanCnt=scanCnt+1,recent=datetime('now','localtime') WHERE openid=?;"
// 	stmtAdd := a.getStmt(runtime, stmtNameAdd)
// 	if stmtAdd == nil {
// 		//lazy setup for stmt
// 		if stmtAdd, err = a.setStmt(runtime, stmtNameAdd, stmtSQLAdd); err != nil {
// 			return
// 		}
// 	}
// 	stmtRecord := a.getStmt(runtime, stmtSQLRecord)
// 	if stmtRecord == nil {
// 		if stmtRecord, err = a.setStmt(runtime, stmtNameRecord, stmtSQLRecord); err != nil {
// 			return
// 		}
// 	}
// 	runtime.mutex.Lock()
// 	defer runtime.mutex.Unlock()
// 	tx, err := runtime.db.Begin()
// 	if err != nil {
// 		return err
// 	}
// 	//Add scan
// 	stmtAdd = tx.Stmt(stmtAdd)
// 	_, err = stmtAdd.Exec(
// 		openid,
// 	)
// 	if err != nil && !strings.HasPrefix(err.Error(), "UNIQUE") {
// 		tx.Rollback()
// 		return err
// 	}
// 	//record scan
// 	stmtRecord = tx.Stmt(stmtRecord)
// 	_, err = stmtRecord.Exec(openid)
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	tx.Commit()
// 	return
// }
