package observers

import (
	"github.com/PharbersDeveloper/bp-jobs-observer/test"
	"testing"
)

func TestBpFileJobsObserver_Open_Exec(t *testing.T) {

	test.SetEnv()

	bfjo := BpFileJobsObserver{
		Id:"test007",
		DBHost:"59.110.31.50",
		DBPort:"5555",
		Database:"pharbers-sandbox-600",
		Collection:"assets",
		Conditions:map[string]interface{}{
			"$and": []map[string]interface{}{
				map[string]interface{}{"file": map[string]interface{}{"$exists": true, "$ne": ""}},
				map[string]interface{}{"isNewVersion": true},
				map[string]interface{}{"dfs": map[string]interface{}{"$exists": true, "$size": 0}},
			},
		},
		ParallelNumber:4,
		RequestTopic:"test006",
		ResponseTopic:"test007",
	}
	bfjo.Open()
	bfjo.Exec()

}
