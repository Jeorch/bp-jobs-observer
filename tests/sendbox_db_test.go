package tests

import (
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-go-lib/redis"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

func TestClearIdInRedis(t *testing.T) {

	SetEnv()

	dbHost := "59.110.31.50"
	dbPort := "5555"
	dbName := "pharbers-sandbox-600"
	dbTable := "datasets"

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:   []string{fmt.Sprintf("%s:%s", dbHost, dbPort)},
		Timeout: 10 * time.Second,
	}

	sess, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		sess.Refresh()
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}
	sess.SetMode(mgo.Monotonic, true)

	var datasets []models.BpDataset
	err = sess.DB(dbName).C(dbTable).Find(bson.M{}).All(&datasets)
	if err != nil {
		panic(err.Error())
	}

	c, err := redis.GetRedisClient()
	if err != nil {
		panic(err.Error())
	}

	var delCount int64

	for _, one := range datasets {
		id := one.Id.Hex()
		i, e := c.Del(id).Result()
		if e != nil {
			fmt.Printf("id=%s del err\n", id)
		}
		if i == 1 {
			delCount += i
		} else {
			fmt.Printf("id=%s del result != 1\n", id)
		}
	}

	fmt.Printf("del count = %d\n", delCount)

}
