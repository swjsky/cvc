1. the rpc logger function can be called by utils.CallRPCLog("addr", logger)
2. below is the test function for calling rpc logger

```
package main

import (
	"loreal.com/dit/projects/logger/modle"
	"loreal.com/dit/utils"
	"fmt"
)

func main() {
	log := &modle.Logger{
		Project: "test_project",
		Method:  "GET",
		Path:    "/user/go",
		Level:   modle.LevelInfo,
		Content: "test content",
	}
	reply, err := utils.CallRPCLog("localhost:1505", log)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(reply)
}

```