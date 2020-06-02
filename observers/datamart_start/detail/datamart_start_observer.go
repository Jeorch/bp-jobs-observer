package detail

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/kafka"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"github.com/PharbersDeveloper/bp-jobs-observer/models/PhEventMsg"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ObserverInfo struct {
	Id             string                 `json:"id"`
	DBHost         string                 `json:"db_host"`
	DBPort         string                 `json:"db_port"`
	Database       string                 `json:"database"`
	Collection     string                 `json:"collection"`
	Conditions     map[string]interface{} `json:"conditions"`
	ParallelNumber int                    `json:"parallel_number"`
	RequestTopic   string                 `json:"request_topic"`
	DBUser         string                 `json:"db_user"`
	DBPass         string                 `json:"db_pass"`
}

var (
	dbSession    *mgo.Session
	kafkaBuilder *kafka.BpKafkaBuilder
	producer     *kafka.BpProducer
	jobChan      chan models.BpDataset
)

func (observer *ObserverInfo) Open() {
	sess, err := mgo.Dial(fmt.Sprintf("%s:%s", observer.DBHost, observer.DBPort))
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}
	cred := mgo.Credential{
		Username: observer.DBUser,
		Password: observer.DBPass,
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
		dbSession = sess
	}

	kafkaBuilder = kafka.NewKafkaBuilder()
	producer, err = kafkaBuilder.BuildProducer()
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}

}

func (observer *ObserverInfo) Exec() {
	execLogger := log.NewLogicLoggerBuilder().SetTraceId(observer.Id).Build()
	execLogger.Info("start exec")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//查询Jobs
	jobs, err := observer.queryJobs()
	if err != nil {
		execLogger.Error(err.Error())
	}

	length := len(jobs)
	execLogger.Info("jobs length=", length)

	if length > 0 {
		jobChan = make(chan models.BpDataset, observer.ParallelNumber)

		//分配worker执行Job
		for id := 1; id <= observer.ParallelNumber; id++ {
			go observer.worker(id, jobChan, ctx)
		}

		//将所有 job push 到 chan队列
		pushJobs(jobChan, jobs)
	}

	execLogger.Info("Ended!")
	time.Sleep(10 * time.Second) //10秒钟后ctx局部变量done
}

func (observer *ObserverInfo) Close() {
	if jobChan != nil {
		close(jobChan)
	}
}

func (observer *ObserverInfo) queryJobs() ([]models.BpDataset, error) {

	logger := log.NewLogicLoggerBuilder().SetTraceId(observer.Id).Build()
	logger.Info("query jobs from db")

	var datasets []models.BpDataset
	//err := dbSession.DB(observer.Database).C(observer.Collection).Find(bson.M{}).All(&datasets)
	err := dbSession.DB(observer.Database).C(observer.Collection).Find(observer.Conditions).All(&datasets)
	if err != nil {
		return nil, err
	}

	var allMart []models.BpMart
	err = dbSession.DB(observer.Database).C("marts").Find(bson.M{}).All(&allMart)
	if err != nil {
		return nil, err
	}
	oldDfsIds, err := observer.getAllMartDfsParents(allMart)
	if err != nil {
		return nil, err
	}

	jobs := make([]models.BpDataset, 0)
	for _, dfs := range datasets {
		if isNotOldDfs(dfs.Id, oldDfsIds) {
			jobs = append(jobs, dfs)
		}
	}

	return jobs, nil
}

func (observer *ObserverInfo) getAllMartDfsParents(allMart []models.BpMart) ([]bson.ObjectId, error) {
	var res []bson.ObjectId
	for _, mart := range allMart {
		for _, id := range mart.Dfs {
			var ds models.BpDataset
			err := dbSession.DB(observer.Database).C(observer.Collection).Find(bson.M{"_id": id}).One(&ds)
			if err != nil {
				return nil, err
			}
			if len(ds.Parent) != 0 {
				res = append(res, ds.Parent...)
			}
		}
	}
	return res, nil
}

func isNotOldDfs(id bson.ObjectId, oldDfsIds []bson.ObjectId) bool {
	for _, oldDfsId := range oldDfsIds {
		if id.Hex() == oldDfsId.Hex() {
			return false
		}
	}
	return true
}

func pushJobs(jobChan chan<- models.BpDataset, jobs []models.BpDataset) {
	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("push jobs to chan")
	for _, job := range jobs {
		jobChan <- job
	}
}

func (observer *ObserverInfo) worker(id int, jobChan <-chan models.BpDataset, ctx context.Context) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	for {
		select {
		case <-ctx.Done():
			workerLogger.Infof("worker-%d stop", id)
			return
		case j := <-jobChan:
			{

				jobLogger := log.NewLogicLoggerBuilder().Build()
				jobLogger.Infof("worker-%d start job=%v", id, j)

				//send job request
				err := observer.sendJobRequest(observer.RequestTopic, j, "append")
				if err != nil {
					jobLogger.Error(err.Error())
				}
				jobLogger.Infof("worker-%d sanded job=%v", id, j.Id.Hex())

				//判断jobChan中是否还有job，若没有，则表示当前job为最后一条，发送end消息
				if len(jobChan) == 0 {
					//send lastJob request
					lastJob := new(models.BpDataset)
					lastJob.Id = bson.NewObjectId()
					err := observer.sendJobRequest(observer.RequestTopic, *lastJob, "end")
					if err != nil {
						jobLogger.Error(err.Error())
					}
					jobLogger.Info("worker-%d sanded lastJob.")
				}
			}
		}
	}

}

func (observer *ObserverInfo) sendJobRequest(topic string, job models.BpDataset, taskType string) error {

	msg := map[string]interface{}{
		"datasetId":job.Id.Hex(),
		"taskType":taskType,
		"url":job.Url,
		"length":job.Length,
		"remarks":"",
	}

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	eventMsg := PhEventMsg.EventMsg{
		JobId:   observer.Id,
		TraceId: observer.Id,
		Type:    "SandBox-hive",
		Data:    string(jsonBytes),
	}

	specificRecordByteArr, err := kafka.EncodeAvroRecord(&eventMsg)
	if err != nil {
		return err
	}

	err = producer.Produce(topic, []byte(observer.Id), specificRecordByteArr)
	return err
}
