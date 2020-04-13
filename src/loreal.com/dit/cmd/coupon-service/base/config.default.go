package base

import(
	"crypto/rsa"
)

//GORoutingNumberForWechat - Total GO Routing # for send process
const GORoutingNumberForWechat = 10

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
var Cfg = Configuration{
	Address:              ":1503",
	Prefix:               "/",
	JwtKey:               "a9ac231b0f2a4f448b8846fd1f57814a",
	AuthPubKey:			  "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxWt9gzvtKVZhN9Xt/t1S5xApkSVeKRDbUGbto1NhWIqZCSgY1bmYVDFgnFLGT9tWdZDR4NMYwJxIpdpqjW/w4Q/4H9ummQE57C/AVQ/d4dJrF6MNyz67TL6kmHnrWCNYdHG9I4buTNCUL2y3DRutZ2nhNED/fDFkvQfWjj0ihqa6+Z4ZVTo0i1pX6u/IAjkHSdFRlzluM9EatuSyPo7T83hYqEjwoXkARLjm9jxPBU9jKOcL/1a3pE1QpTisxiQeIsmcbzRH/DPOhbJUwueQ3ux1CGu9RDZ8AX8eZvTrvXF41/b7N4cOi5jUvmV2H02NQh7WLp60Ln/hYmf5+nV5UwIDAQAB\n-----END PUBLIC KEY-----",	
	AppTitle:             "Loreal coupon service",
	Production:           false,
	Envs: []*Env{
		{
			Name:       "prod",
			SqliteDB:   "data.db",
			DataFolder: "./data/",
		},
	},
	ScheduledTasks: []*ScheduledTask{
		{Schedule: "0 0 0 * * *", Task: "daily-maintenance", DefaultArgs: []string{}},
		{Schedule: "0 10 0 * * *", Task: "daily-maintenance-pp", DefaultArgs: []string{}},
	},
}

var Pubkey *rsa.PublicKey