package clean_err_dfs

import (
	"github.com/PharbersDeveloper/bp-jobs-observer/observers"
	clean_err_dfs "github.com/PharbersDeveloper/bp-jobs-observer/observers/clean_err_dfs/detail"
	"github.com/hashicorp/go-uuid"
	"os"
	"strconv"
)

const (
	EntryValue = "clean-err-dfs-start"
	AwsRegion = "AWS_REGION"
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

	newId, _ := uuid.GenerateUUID()
	//TODO: Conditions 配置抽离
	bpjo := clean_err_dfs.ObserverInfo{
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
	}
	bpjo.Open()
	bpjo.Exec()
	bpjo.Close()
}


