package tests

import (
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestDelete(t *testing.T) {

	sess, err := mgo.Dial(fmt.Sprintf("%s:%s", "192.168.100.116", "27017"))
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}
	cred := mgo.Credential{
		Username: "",
		Password: "",
	}
	err = sess.Login(&cred)
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
		if sess != nil {
			sess.Refresh()
		}
	}
	if sess != nil {
		sess.SetMode(mgo.Monotonic, true)
	}
	logger := log.NewLogicLoggerBuilder().Build()
	err = sess.DB("pharbers-sandbox-merge2").C("jeo-test-dfs").RemoveId(bson.ObjectIdHex("5ecf5b26b958e100019483bd"))
	if err != nil {
		logger.Error(err.Error())
	}

}

func TestSearch(t *testing.T) {

	sess, err := mgo.Dial(fmt.Sprintf("%s:%s", "192.168.100.116", "27017"))
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}
	cred := mgo.Credential{
		Username: "",
		Password: "",
	}
	err = sess.Login(&cred)
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
		if sess != nil {
			sess.Refresh()
		}
	}
	if sess != nil {
		sess.SetMode(mgo.Monotonic, true)
	}
	logger := log.NewLogicLoggerBuilder().Build()
	dfsId := bson.ObjectIdHex("5ecf5b26b958e100019483bd")

	var asset models.BpAsset
	err = sess.DB("pharbers-sandbox-merge2").C("jeo-test-asset").Find(bson.M{"dfs":bson.M{"$in":[]bson.ObjectId{dfsId}}}).One(&asset)
	if err != nil {
		logger.Error(err.Error())
	}
	fmt.Println(asset)

	if len(asset.Dfs) != 0 {
		asset.Dfs = nil
		err = sess.DB("pharbers-sandbox-merge2").C("jeo-test-asset").UpdateId(asset.Id, asset)
		if err != nil {
			logger.Error(err.Error())
		}
		fmt.Println(asset)
	}


}
