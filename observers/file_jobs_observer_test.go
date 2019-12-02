package observers

import (
	"context"
	"fmt"
	"github.com/PharbersDeveloper/bp-jobs-observer/test"
	"testing"
)

func TestBpFileJobsObserver_Open_Exec(t *testing.T) {

	test.SetEnv()

	bfjo := BpFileJobsObserver{
		Id:         "test007",
		DBHost:     "59.110.31.50",
		DBPort:     "5555",
		Database:   "pharbers-sandbox-600",
		Collection: "assets",
		Conditions: map[string]interface{}{
			"$and": []map[string]interface{}{
				map[string]interface{}{"file": map[string]interface{}{"$exists": true, "$ne": ""}},
				map[string]interface{}{"isNewVersion": true},
				map[string]interface{}{"dfs": map[string]interface{}{"$exists": true, "$size": 0}},
			},
		},
		ParallelNumber:         1,
		SingleJobTimeoutSecond: 10,
		ScheduleDurationSecond: 60,
		RequestTopic:           "test006",
		ResponseTopic:          "test007",
	}
	bfjo.Open()
	bfjo.Exec()

}

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	//defer func() {
	//	if r := recover(); r != nil {
	//		fmt.Println("Recovered in f", r)
	//	}
	//}()

	go p(ctx)

	panic("aaa")

	select {
	case <-ctx.Done():
		fmt.Println("main", ctx.Err())
	}


}

func p(ctx context.Context)  {

	fmt.Println("ppp")
	select {
	case <-ctx.Done():
		fmt.Println("handle", ctx.Err())
	}
	fmt.Println("ppp done")

}

func TestReSetMaxConf(t *testing.T) {

}
