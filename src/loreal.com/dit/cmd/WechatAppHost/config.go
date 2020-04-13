package main

import "strings"

//Config - data struct for configuration file
type Config struct {
	AppID                string           `json:"appid"`
	Address              string           `json:"address"`
	Prefix               string           `json:"prefix"`
	RedisServerStr       string           `json:"redis-server"`
	AppDomainName        string           `json:"app-domain-name"`
	TokenServiceURL      string           `json:"token-service-url"`
	TokenServiceUsername string           `json:"token-service-user"`
	TokenServicePassword string           `json:"token-service-password"`
	Envs                 []*Env           `json:"envs,omitempty"`
	ScheduledTasks       []*ScheduledTask `json:"scheduled-tasks,omitempty"`
}

func (c *Config) fixPrefix() {
	if !strings.HasPrefix(c.Prefix, "/") {
		c.Prefix = "/" + c.Prefix
	}
	if !strings.HasSuffix(c.Prefix, "/") {
		c.Prefix = c.Prefix + "/"
	}
}

//Env - env configuration
type Env struct {
	Name       string `json:"name,omitempty"`
	SqliteDB   string `json:"sqlite-db,omitempty"`
	DataFolder string `json:"data,omitempty"`
}

//ScheduledTask - command and parameters to run
/* Schedule Spec:
Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight on Sunday        | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *

***
*** corn example ***:

c := cron.New()
c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
c.AddFunc("@hourly",      func() { fmt.Println("Every hour") })
c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty") })
c.Start()
..
// Funcs are invoked in their own goroutine, asynchronously.
...
// Funcs may also be added to a running Cron
c.AddFunc("@daily", func() { fmt.Println("Every day") })
..
// Inspect the cron job entries' next and previous run times.
inspect(c.Entries())
..
c.Stop()  // Stop the scheduler (does not stop any jobs already running).

*/
//ScheduledTask - Scheduled Task
type ScheduledTask struct {
	Schedule    string   `json:"schedule,omitempty"`
	Task        string   `json:"task,omitempty"`
	DefaultArgs []string `json:"default-args,omitempty"`
}
