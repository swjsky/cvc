package main

import "loreal.com/dit/utils/task"

func (a *App) registerTasks() {
	a.TaskManager.RegisterWithContext("daily-maintenance-pp", "ceh-cs-test", a.dailyMaintenanceTaskHandler, 1)
	a.TaskManager.RegisterWithContext("daily-maintenance", "ceh-cs", a.dailyMaintenanceTaskHandler, 1)
}

//dailyMaintenanceTaskHandler - run daily maintenance task
func (a *App) dailyMaintenanceTaskHandler(t *task.Task, args ...string) {
	//a.DailyMaintenance(t, task.GetArgs(args, 0))
	a.DailyMaintenance(t)
}
