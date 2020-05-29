package datamart_start

import (
	"github.com/PharbersDeveloper/bp-jobs-observer/observers"
	datamart "github.com/PharbersDeveloper/bp-jobs-observer/observers/datamart_start/detail"
	"github.com/hashicorp/go-uuid"
	"os"
	"strconv"
)

const (
	EntryValue = "datamart-start"
)

func Run() {

	DbHost := os.Getenv(observers.DbHostKey)
	if DbHost == "" {
		println("Error! No DB_HOST env set.")
		return
	}
	DbPort := os.Getenv(observers.DbPortKey)
	if DbPort == "" {
		println("Error! No DB_PORT env set.")
		return
	}
	DbUser := os.Getenv(observers.DbUserKey)
	if DbUser == "" {
		println("Warn! No DB_USER env set.")
	}
	DbPass := os.Getenv(observers.DbPassKey)
	if DbPass == "" {
		println("Warn! No DB_PASS env set.")
	}
	DbName := os.Getenv(observers.DbNameKey)
	if DbName == "" {
		println("Error! No DB_NAME env set.")
		return
	}
	DbColl := os.Getenv(observers.DbCollKey)
	if DbColl == "" {
		println("Error! No DB_COLL env set.")
		return
	}
	ParallelNumStr := os.Getenv(observers.ParallelNumKey)
	if ParallelNumStr == "" {
		println("Error! No PARALLEL_NUM env set.")
		return
	}
	ParallelNum, err := strconv.Atoi(ParallelNumStr)
	if err != nil {
		panic(err.Error())
	}
	ReqTopic := os.Getenv(observers.ReqTopicKey)
	if ReqTopic == "" {
		println("Error! No REQ_TOPIC env set.")
		return
	}

	newId, _ := uuid.GenerateUUID()
	//TODO: Conditions 配置抽离
	bpjo := datamart.ObserverInfo{
		Id:         newId,
		DBHost:     DbHost,
		DBPort:     DbPort,
		DBUser:     DbUser,
		DBPass:     DbPass,
		Database:   DbName,
		Collection: DbColl,
		Conditions: map[string]interface{}{
			"$and": []map[string]interface{}{
				map[string]interface{}{"url": map[string]interface{}{"$exists": true, "$ne": ""}},
				map[string]interface{}{"status": "end"},
				map[string]interface{}{"description": "pyJob"},
			},
		},
		ParallelNumber: ParallelNum,
		RequestTopic:   ReqTopic,
	}
	bpjo.Open()
	bpjo.Exec()
	bpjo.Close()
}

//本地开发测试使用
func SetEntryEnv() {

	_ = os.Setenv(observers.EntryKey, EntryValue)
	_ = os.Setenv(observers.DbHostKey, "192.168.100.116")
	_ = os.Setenv(observers.DbPortKey, "27017")
	_ = os.Setenv(observers.DbUserKey, "")
	_ = os.Setenv(observers.DbPassKey, "")
	_ = os.Setenv(observers.DbNameKey, "pharbers-sandbox-merge2")
	_ = os.Setenv(observers.DbCollKey, "datasets")
	_ = os.Setenv(observers.ParallelNumKey, "1")
	_ = os.Setenv(observers.ReqTopicKey, "test528")

}
