package main

import (
	"strings"
)

func (a *App) getUserID(
	runtime *RuntimeEnv,
	UID string,
	userID *int64,
) (err error) {
	*userID = -1
	const stmtNameGet = "GetUserID"
	const stmtSQLGet = "SELECT ID FROM User WHERE UID=?;"
	stmtGet := a.getStmt(runtime, stmtNameGet)
	if stmtGet == nil {
		//lazy setup for stmt
		if stmtGet, err = a.setStmt(runtime, stmtNameGet, stmtSQLGet); err != nil {
			return
		}
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	//get customerID
	if err = stmtGet.QueryRow(
		UID,
	).Scan(userID); err != nil {
		return err
	}
	return
}

func (a *App) recordWxUser(
	runtime *RuntimeEnv,
	openID, nickName, avatar, scene, pageID string,
	customerID *int64,
) (err error) {
	*customerID = -1
	const stmtNameAdd = "AddUser"
	const stmtSQLAdd = "INSERT INTO User (OpenID,NickName,Avatar,Scene,CreateAt) VALUES (?,?,?,?,datetime('now','localtime'));"
	const stmtNameGet = "GetWxUserID"
	const stmtSQLGet = "SELECT ID FROM User WHERE OpenID=?;"
	const stmtNameRecord = "RecordVisit"
	const stmtSQLRecord = "UPDATE User SET PV=PV+1 WHERE ID=?;"
	const stmtNameNewPV = "InsertVisit"
	const stmtSQLNewPV = "INSERT INTO Visit (UserID,PageID,Scene,CreateAt,Recent,PV,State) VALUES (?,?,?,datetime('now','localtime'),datetime('now','localtime'),1,1);"
	const stmtNamePV = "RecordPV"
	const stmtSQLPV = "UPDATE Visit SET PV=PV+1,Recent=datetime('now','localtime'),State=1 WHERE WxUserID=? AND PageID=? AND Scene=?;"
	stmtAdd := a.getStmt(runtime, stmtNameAdd)
	if stmtAdd == nil {
		//lazy setup for stmt
		if stmtAdd, err = a.setStmt(runtime, stmtNameAdd, stmtSQLAdd); err != nil {
			return
		}
	}
	stmtGet := a.getStmt(runtime, stmtNameGet)
	if stmtGet == nil {
		//lazy setup for stmt
		if stmtGet, err = a.setStmt(runtime, stmtNameGet, stmtSQLGet); err != nil {
			return
		}
	}
	stmtRecord := a.getStmt(runtime, stmtNameRecord)
	if stmtRecord == nil {
		if stmtRecord, err = a.setStmt(runtime, stmtNameRecord, stmtSQLRecord); err != nil {
			return
		}
	}
	stmtNewPV := a.getStmt(runtime, stmtNameNewPV)
	if stmtNewPV == nil {
		if stmtNewPV, err = a.setStmt(runtime, stmtNameNewPV, stmtSQLNewPV); err != nil {
			return
		}
	}
	stmtPV := a.getStmt(runtime, stmtNamePV)
	if stmtPV == nil {
		if stmtPV, err = a.setStmt(runtime, stmtNamePV, stmtSQLPV); err != nil {
			return
		}
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	tx, err := runtime.db.Begin()
	if err != nil {
		return err
	}
	//Add user
	stmtAdd = tx.Stmt(stmtAdd)
	result, err := stmtAdd.Exec(
		openID,
		nickName,
		avatar,
		scene,
	)
	if err != nil && !strings.HasPrefix(err.Error(), "UNIQUE") {
		tx.Rollback()
		return err
	}
	//get customerID
	if result == nil {
		//find user
		stmtGet = tx.Stmt(stmtGet)
		if err = stmtGet.QueryRow(
			openID,
		).Scan(customerID); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		*customerID, err = result.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	//record customer visit
	//log.Println("[INFO] - Add user:", *customerID)
	stmtRecord = tx.Stmt(stmtRecord)
	_, err = stmtRecord.Exec(*customerID)
	if err != nil {
		tx.Rollback()
		return err
	}
	stmtPV = tx.Stmt(stmtPV)
	pvResult, err := stmtPV.Exec(
		*customerID,
		pageID,
		scene,
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
		*customerID,
		pageID,
		scene,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return
}

func (a *App) recordPV(
	runtime *RuntimeEnv,
	customerID int64,
	pageID, scene string,
	visitState int,
) (err error) {
	const stmtNameNewPV = "InsertVisit1"
	const stmtSQLNewPV = "INSERT INTO Visit (CustomerID,PageID,Scene,CreateAt,Recent,PV,State) VALUES (?,?,?,datetime('now','localtime'),datetime('now','localtime'),1,?);"
	const stmtNamePV = "UpdatePV"
	const stmtSQLPV = "UPDATE Visit SET PV=PV+1,Recent=datetime('now','localtime') WHERE CustomerID=? AND PageID=? AND Scene=? AND State=?;"
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
		customerID,
		pageID,
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
		customerID,
		pageID,
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
