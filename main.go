package main

import (
	"github.com/PharbersDeveloper/bp-jobs-observer/observers"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/clean_err_dfs"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/datamart_start"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/oss_task_start"
	"log"
	"os"
)

func main() {

	//本地开发调试使用，部署时请注释掉下面 setDevEnv 行
	//setDevEnv(clean_err_dfs.EntryValue)

	entry := os.Getenv(observers.EntryKey)
	if entry == "" {
		println("Error! No BP_JOBS_OBSERVER_ENTRY env set.")
		return
	}

	//TODO：服务注册
	switch entry {
	case datamart_start.EntryValue:
		datamart_start.Run()
	case oss_task_start.EntryValue:
		oss_task_start.Run()
	case clean_err_dfs.EntryValue:
		clean_err_dfs.Run()
	default:
		log.Fatalf("Error! Env BP_JOBS_OBSERVER_ENTRY(%s) not implemented.", entry)
	}

}
