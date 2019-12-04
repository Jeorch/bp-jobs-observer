// Package test is bp-go-lib's common test constants or functions.
package tests

import (
	"github.com/PharbersDeveloper/bp-go-lib/env"
	"os"
)

func SetEnv() {
	//项目范围内的环境变量
	_ = os.Setenv(env.ProjectName, "bp-jobs-observer")

	//log
	_ = os.Setenv(env.LogTimeFormat, "2006-01-02 15:04:05")
	_ = os.Setenv(env.LogOutput, "console")
	//_ = os.Setenv(env.LogOutput, "./tmp/bp-jobs-observer.log")
	_ = os.Setenv(env.LogLevel, "info")

	//kafka
	_ = os.Setenv(env.KafkaConfigPath, "../resources/kafka_config.json")
	_ = os.Setenv(env.KafkaSchemaRegistryUrl, "http://123.56.179.133:8081")

	//redis
	_ = os.Setenv(env.RedisHost, "192.168.100.176")
	_ = os.Setenv(env.RedisPort, "6379")
	_ = os.Setenv(env.RedisPass, "")
	_ = os.Setenv(env.RedisDb, "0")

}
