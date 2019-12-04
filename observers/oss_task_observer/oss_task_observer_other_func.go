package oss_task_observer

import (
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"gopkg.in/mgo.v2/bson"
)

//压测使用
func (observer *ObserverInfo) updateDfs(jobId string) {

	//TODO: 由于asset是以traceId为区分，jobId未使用，这里自动化Job以traceId作为JobId
	traceId := jobId
	logger := log.NewLogicLoggerBuilder().SetJobId(jobId).SetTraceId(traceId).Build()
	var asset models.BpAsset
	err := dbSession.DB(observer.Database).C(observer.Collection).Find(bson.M{"traceId": traceId}).One(&asset)
	if err != nil {
		logger.Error(err.Error())
	}

	asset.Dfs = nil

	err = dbSession.DB(observer.Database).C(observer.Collection).UpdateId(asset.Id, asset)
	if err != nil {
		logger.Error(err.Error())
	}
}
