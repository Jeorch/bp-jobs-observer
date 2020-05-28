package tests

import (
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

func TestReadChartInDB(t *testing.T) {

	SetEnv()

	dbHost := "59.110.31.50"
	dbPort := "5555"
	dbName := "pharbers-dashboard-config"
	dbTable := "Chart"
	id := bson.ObjectIdHex("5e15ca980f84221d01409e63")

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:   []string{fmt.Sprintf("%s:%s", dbHost, dbPort)},
		Timeout: 20 * time.Second,
	}

	sess, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		sess.Refresh()
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}
	sess.SetMode(mgo.Monotonic, true)

	var chart models.Chart
	err = sess.DB(dbName).C(dbTable).Find(bson.M{"_id": id}).One(&chart)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(chart.Id)
	fmt.Println(chart.MetadataId)
	fmt.Println(chart.StyleConfigs)
	fmt.Println(chart.DataConfigs)

}
