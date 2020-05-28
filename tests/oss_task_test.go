package tests

import (
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/oss_task_start/detail"
	"github.com/hashicorp/go-uuid"
	"testing"
)

func TestOssTask(t *testing.T) {

	SetEnv()

	newId, _ := uuid.GenerateUUID()
	bpjo := detail.ObserverInfo{
		Id:         newId,
		DBHost:     "59.110.31.50",
		DBPort:     "5555",
		Database:   "pharbers-sandbox-merge",
		Collection: "assets",
		Conditions: map[string]interface{}{
			"$and": []map[string]interface{}{
				map[string]interface{}{"file": map[string]interface{}{"$exists": true, "$ne": ""}},
				map[string]interface{}{"isNewVersion": true},
				map[string]interface{}{"dfs": map[string]interface{}{"$exists": true, "$size": 0}},
			},
		},
		ParallelNumber: 1,
		RequestTopic:   "test001",
	}
	bpjo.Open()
	bpjo.Exec()
	bpjo.Close()

}
