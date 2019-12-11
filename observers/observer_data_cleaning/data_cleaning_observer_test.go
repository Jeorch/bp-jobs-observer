package observer_data_cleaning

import (
	"github.com/PharbersDeveloper/bp-go-lib/env"
	"os"
	"testing"
)

func TestStartDataCleanObserver(t *testing.T) {
	setEnv()

	bfjo := ObserverInfo{
		Id:         "0000001",
		DBHost:     "59.110.31.50",
		DBPort:     "5555",
		Database:   "pharbers-sandbox-600",
		Collection: "datasets",
		Conditions: map[string]interface{}{
			"$and": []map[string]interface{}{
				map[string]interface{}{"url": map[string]interface{}{"$exists": true, "$ne": ""}},
				map[string]interface{}{"description": "Python 清洗 Job"},
			},
		},
		ParallelNumber:         1,
		SingleJobTimeoutSecond: 3600,
		ScheduleDurationSecond: 3600,
		RequestTopic:           "HiveTask",
		ResponseTopic:          "HiveTaskResponse",
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
	_ = os.Setenv(env.KafkaConfigPath, "../../resources/kafka_config.json")
	_ = os.Setenv(env.KafkaSchemaRegistryUrl, "http://123.56.179.133:8081")

	//redis
	_ = os.Setenv(env.RedisHost, "192.168.100.176")
	_ = os.Setenv(env.RedisPort, "6379")
	_ = os.Setenv(env.RedisPass, "")
	_ = os.Setenv(env.RedisDb, "0")
}
