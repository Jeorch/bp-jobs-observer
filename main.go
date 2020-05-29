package main

import (
	"github.com/PharbersDeveloper/bp-go-lib/env"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/datamart_start"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/oss_task_start"
	"log"
	"os"
)

func main() {

	//本地开发调试使用，部署时请注释掉下面 setDevEnv 行
	//setDevEnv(datamart_start.EntryValue)

	entry := os.Getenv(observers.EntryKey)
	if entry == "" {
		println("Error! No BP_JOBS_OBSERVER_ENTRY env set.")
		return
	}

	switch entry {
	case datamart_start.EntryValue:
		datamart_start.Run()
	case oss_task_start.EntryValue:
		oss_task_start.Run()
	default:
		log.Fatalf("Error! Env BP_JOBS_OBSERVER_ENTRY(%s) not implemented.", entry)
	}

}

func setDevEnv(entry string) {
	//项目范围内的环境变量
	_ = os.Setenv(env.ProjectName, "bp-jobs-servers")

	//Entry内的环境变量
	switch entry {
	case datamart_start.EntryValue:
		datamart_start.SetEntryEnv()
	case oss_task_start.EntryValue:
		oss_task_start.SetEntryEnv()
	default:
		log.Fatalf("Error! Env BP_JOBS_OBSERVER_ENTRY(%s) not implemented.", entry)
	}

	//log
	_ = os.Setenv(env.LogTimeFormat, "2006-01-02 15:04:05")
	_ = os.Setenv(env.LogOutput, "console")
	//_ = os.Setenv(env.LogOutput, "./logs/bp-jobs-observer.log")
	_ = os.Setenv(env.LogLevel, "info")

	//kafka
	_ = os.Setenv(env.KafkaConfigPath, "deploy-config/kafka_config.json")
	_ = os.Setenv(env.KafkaSchemaRegistryUrl, "http://schema.message:8081")
}
