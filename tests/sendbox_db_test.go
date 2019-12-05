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

func TestClearAssetIdInRedis(t *testing.T) {

	SetEnv()

	dbHost := "59.110.31.50"
	dbPort := "5555"
	dbName := "pharbers-sandbox-600"
	dbTable := "assets"

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

	var assets []models.BpAsset
	err = sess.DB(dbName).C(dbTable).Find(bson.M{}).All(&assets)
	if err != nil {
		panic(err.Error())
	}

	c, err := redis.GetRedisClient()
	if err != nil {
		panic(err.Error())
	}

	for _, asset := range assets {
		assetId := asset.Id.Hex()
		i, e := c.Del(assetId).Result()
		if e != nil {
			fmt.Printf("assetId=%s del err\n", assetId)
		}
		if i != 1 {
			fmt.Printf("assetId=%s del result != 1\n", assetId)
		}
	}

}
