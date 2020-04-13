package main

import (
	"loreal.com/dit/utils/task"
)

//DailyMaintenance - task to do daily maintenance
func (a *App) DailyMaintenance(t *task.Task) (err error) {
	// const stmtName = "dm-clean-vehicle"
	// const stmtSQL = "DELETE FROM vehicle_left WHERE enter<=?;"
	// env := getEnv(t.Context)
	// runtime := a.getRuntime(env)
	// if runtime == nil {
	// 	return ErrMissingRuntime
	// }
	// stmt := a.getStmt(runtime, stmtName)
	// if stmt == nil {
	// 	//lazy setup for stmt
	// 	if stmt, err = a.setStmt(runtime, stmtName, stmtSQL); err != nil {
	// 		return err
	// 	}
	// }
	// runtime.mutex.Lock()
	// defer runtime.mutex.Unlock()
	// _, err = stmt.Exec(int(time.Now().Add(time.Hour * -168).Unix())) /* 7*24Hours = 168*/
	return nil
}
