package main

import (
	"github.com/PharbersDeveloper/bp-go-lib/env"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/observer_oss_task"
	"github.com/hashicorp/go-uuid"
	"os"
	"strconv"
)

const (
	DbHostKey      = "DB_HOST"
	DbPortKey      = "DB_PORT"
	DbNameKey      = "DB_NAME"
	DbCollKey      = "DB_COLL"
	ParallelNumKey = "PARALLEL_NUM"
	ReqTopicKey    = "REQ_TOPIC"
)

func main() {
	//本机调试使用，部署时请注释掉下面setEnv行
	//setEnv()

	DbHost := os.Getenv(DbHostKey)
	if DbHost == "" {
		println("Error! No DB_HOST env set.")
		return
	}
	DbPort := os.Getenv(DbPortKey)
	if DbPort == "" {
		println("Error! No DB_PORT env set.")
		return
	}
	DbName := os.Getenv(DbNameKey)
	if DbName == "" {
		println("Error! No DB_NAME env set.")
		return
	}
	DbColl := os.Getenv(DbCollKey)
	if DbColl == "" {
		println("Error! No DB_COLL env set.")
		return
	}
	ParallelNumStr := os.Getenv(ParallelNumKey)
	if ParallelNumStr == "" {
		println("Error! No PARALLEL_NUM env set.")
		return
	}
	ParallelNum, err := strconv.Atoi(ParallelNumStr)
	if err != nil {
		panic(err.Error())
	}
	ReqTopic := os.Getenv(ReqTopicKey)
	if ReqTopic == "" {
		println("Error! No REQ_TOPIC env set.")
		return
	}

	newId, _ := uuid.GenerateUUID()
	//TODO: Conditions 配置抽离
	bpjo := observer_oss_task.ObserverInfo{
		Id:         newId,
		DBHost:     DbHost,
		DBPort:     DbPort,
		Database:   DbName,
		Collection: DbColl,
		Conditions: map[string]interface{}{
			"$and": []map[string]interface{}{
				map[string]interface{}{"file": map[string]interface{}{"$exists": true, "$ne": ""}},
				map[string]interface{}{"isNewVersion": true},
				map[string]interface{}{"dfs": map[string]interface{}{"$exists": true, "$size": 0}},
			},
		},
		ParallelNumber: ParallelNum,
		RequestTopic:   ReqTopic,
	}
	bpjo.Open()
	bpjo.Exec()
	bpjo.Close()
}

func setEnv() {
	//项目范围内的环境变量
	_ = os.Setenv(env.ProjectName, "bp-jobs-observer")
	_ = os.Setenv(DbHostKey, "127.0.0.1")
	_ = os.Setenv(DbPortKey, "27017")
	_ = os.Setenv(DbNameKey, "pharbers-sandbox-merge")
	_ = os.Setenv(DbCollKey, "assets")
	_ = os.Setenv(ParallelNumKey, "1")
	_ = os.Setenv(ReqTopicKey, "oss_msg")

	//log
	_ = os.Setenv(env.LogTimeFormat, "2006-01-02 15:04:05")
	//_ = os.Setenv(env.LogOutput, "console")
	_ = os.Setenv(env.LogOutput, "./logs/bp-jobs-observer.log")
	_ = os.Setenv(env.LogLevel, "info")

	//kafka
	_ = os.Setenv(env.KafkaConfigPath, "resources/kafka_config.json")
	_ = os.Setenv(env.KafkaSchemaRegistryUrl, "http://schema.message:8081")

	//redis
	_ = os.Setenv(env.RedisHost, "192.168.100.176")
	_ = os.Setenv(env.RedisPort, "6379")
	_ = os.Setenv(env.RedisPass, "")
	_ = os.Setenv(env.RedisDb, "0")
}
