package main

import (
	"github.com/PharbersDeveloper/bp-go-lib/env"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/observer_oss_task"
	"os"
)

func main() {
	setEnv()

	bfjo := observer_oss_task.ObserverInfo{
		Id:         "0000001",
		DBHost:     "59.110.31.50",
		DBPort:     "5555",
		Database:   "pharbers-sandbox-600-chc",
		Collection: "assets",
		Conditions: map[string]interface{}{
			"$and": []map[string]interface{}{
				map[string]interface{}{"file": map[string]interface{}{"$exists": true, "$ne": ""}},
				map[string]interface{}{"isNewVersion": true},
				map[string]interface{}{"dfs": map[string]interface{}{"$exists": true, "$size": 0}},
			},
		},
		ParallelNumber:         1,
		SingleJobTimeoutSecond: 3600,
		ScheduleDurationSecond: 3600,
		RequestTopic:           "oss_task_submit",
		ResponseTopic:          "oss_task_response",
	}
	bfjo.Open()
	bfjo.Exec()
}

func setEnv() {
	//项目范围内的环境变量
	_ = os.Setenv(env.ProjectName, "bp-jobs-observer")

	//log
	_ = os.Setenv(env.LogTimeFormat, "2006-01-02 15:04:05")
	//_ = os.Setenv(env.LogOutput, "console")
	_ = os.Setenv(env.LogOutput, "./logs/bp-jobs-observer.log")
	_ = os.Setenv(env.LogLevel, "info")

	//kafka
	_ = os.Setenv(env.KafkaConfigPath, "resources/kafka_config.json")
	_ = os.Setenv(env.KafkaSchemaRegistryUrl, "http://123.56.179.133:8081")

	//redis
	_ = os.Setenv(env.RedisHost, "192.168.100.176")
	_ = os.Setenv(env.RedisPort, "6379")
	_ = os.Setenv(env.RedisPass, "")
	_ = os.Setenv(env.RedisDb, "0")
}
