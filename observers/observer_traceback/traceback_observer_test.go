package observer_traceback

import (
	"bufio"
	"github.com/PharbersDeveloper/bp-go-lib/env"
	"github.com/PharbersDeveloper/bp-jobs-observer/utils"
	"os"
	"testing"
)

func TestStartTracebackObserver(t *testing.T) {
	setEnv()

	bfjo := ObserverInfo{
		Id:                     "0000003",
		DBHost:                 "59.110.31.50",
		DBPort:                 "5555",
		Database:               "pharbers-sandbox-600",
		Collection:             "datasets",
		Conditions:             map[string]interface{}{},
		ParallelNumber:         1,
		SingleJobTimeoutSecond: 60,
		ScheduleDurationSecond: 3600,
		RequestTopic:           "HiveTracebackTask",
		ResponseTopic:          "HiveTracebackTaskResponse",
		FilePath:               "../../resources/tmp/error4",
	}
	bfjo.Open()
	bfjo.Exec()
}

func setEnv() {
	//项目范围内的环境变量
	_ = os.Setenv(env.ProjectName, "bp-jobs-observer")

	//log
	_ = os.Setenv(env.LogTimeFormat, "2006-01-02 15:04:05")
	_ = os.Setenv(env.LogOutput, "console")
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

func TestReadFile(t *testing.T) {
	filePath := "../../resources/tmp/error"

	//dat, err := ioutil.ReadFile(filePath)
	//utils.Check(err)
	//allData := string(dat)
	////fmt.Print(allData)
	//
	//arr := strings.Split(allData, "\n")
	//
	//fmt.Println(len(arr))
	//for _, v := range arr {
	//	fmt.Println(v)
	//}

	file, err := os.Open(filePath)
	if err != nil {
		utils.Check(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var arr []interface{}
	for scanner.Scan() {
		arr = append(arr, scanner.Text())
		//fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		utils.Check(err)
	}
}
