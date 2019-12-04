package oss_task_observer

import (
	"context"
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"github.com/hashicorp/go-uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

func TestMgoDialInfo(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	go func(ctx context.Context) {
		for i := 1; i < 20; i++ {
			fmt.Printf("第%d次\n", i)

			var asset models.BpAsset
			err = sess.DB(dbName).C(dbTable).Find(bson.M{}).Skip(i).One(&asset)
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("assetId=%s\n", asset.Id.Hex())

			time.Sleep(time.Second)
		}
		select {
		case <-ctx.Done():
			fmt.Println("End")
		}
	}(ctx)

	select {
	case <-ctx.Done():
		fmt.Println("End")
	}

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

func p(ctx context.Context) {

	fmt.Println("ppp")
	select {
	case <-ctx.Done():
		fmt.Println("handle", ctx.Err())
	}
	fmt.Println("ppp done")

}

func TestUUID(t *testing.T) {

	id, err := uuid.GenerateUUID()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(id)

}
