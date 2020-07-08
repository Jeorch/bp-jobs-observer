package detail

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/kafka"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"github.com/PharbersDeveloper/bp-jobs-observer/models/PhEventMsg"
	"github.com/PharbersDeveloper/bp-jobs-observer/models/record"
	"github.com/hashicorp/go-uuid"
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
	jobChan      chan record.OssTask
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
		jobChan = make(chan record.OssTask, observer.ParallelNumber)

		//分配worker执行Job
		for id := 1; id <= observer.ParallelNumber; id++ {
			go observer.worker(id, jobChan, ctx)
		}

		//将一组 job push 到 chan队列
		pushJobs(jobChan, jobs[0:observer.ParallelNumber])
	}

	execLogger.Info("Oss Task Ended!")
	time.Sleep(10 * time.Second) //10秒钟后ctx局部变量done
}

func (observer *ObserverInfo) Close() {
	if jobChan != nil {
		close(jobChan)
	}
}

func (observer *ObserverInfo) queryJobs() ([]record.OssTask, error) {

	logger := log.NewLogicLoggerBuilder().SetTraceId(observer.Id).Build()
	logger.Info("query jobs from db")

	var assets []models.BpAsset
	err := dbSession.DB(observer.Database).C(observer.Collection).Find(observer.Conditions).All(&assets)
	if err != nil {
		return nil, err
	}

	jobs := make([]record.OssTask, 0)
	for _, asset := range assets {

		//if len(asset.Providers) != 2 || asset.Providers[1] == "CHC" {
		//	continue
		//}

		file, e := observer.queryFile(asset.File)
		if e != nil {
			logger.Error(e.Error())
			continue
		}
		if file.Label != "原始数据" {
			continue
		}

		if file.Extension == "xlsx" || file.Extension == "xls" || file.Extension == "csv" {
			assetId := asset.Id.Hex()
			newId, _ := uuid.GenerateUUID()
			//providers, err := getProviders(asset)
			//providers, err := convertProviders(asset)
			if err != nil {
				logger.Errorf("Get providers'error: %s.", err.Error())
			}

			job := record.OssTask{
				AssetId:    assetId,
				JobId:      newId,
				TraceId:    observer.Id,
				OssKey:     file.Url,
				FileType:   file.Extension,
				FileName:   file.FileName,
				SheetName:  file.SheetName,
				Owner:      asset.Owner,
				CreateTime: int64(asset.CreateTime),
				Labels:     asset.Labels,
				DataCover:  asset.DataCover,
				GeoCover:   asset.GeoCover,
				Markets:    asset.Markets,
				Molecules:  asset.Molecules,
				Providers:  asset.Providers,
			}
			jobs = append(jobs, job)
		} else {
			logger.Warnf("not match file extension(%s)", file.Extension)
		}

	}

	return jobs, nil
}

func (observer *ObserverInfo) queryFile(id bson.ObjectId) (models.BpFile, error) {
	var file models.BpFile
	err := dbSession.DB(observer.Database).C("files").Find(bson.M{"_id": id}).One(&file)
	if err != nil {
		return file, err
	}
	return file, err
}

func convertProviders(asset models.BpAsset) ([]string, error) {
	switch asset.Providers[1] {
	case "CPA&PTI&DDD&HH", "CPA", "GYC", "GYC&CPA":
		asset.Providers[1] = "CPA&GYC"
	}
	return asset.Providers, nil
}

func getProviders(asset models.BpAsset) ([]string, error) {

	if asset.Labels == nil {
		return asset.Providers, nil
	}

	//TODO：听说之后labels这个会变，目前是 json-string-array
	for _, label := range asset.Labels {
		m := make(map[string]interface{}, 0)
		err := json.Unmarshal([]byte(label), &m)
		if err != nil {
			return asset.Providers, err
		}
		if pInterface, ok := m["providers"]; ok {
			pArr, ok := pInterface.([]interface{})
			if !ok {
				return asset.Providers, nil
			}
			providers := make([]string, 0)
			for _, p := range pArr {
				pM, ok := p.(map[string]interface{})
				if ok && pM != nil {
					for _, v := range pM {
						if pStr, ok := v.(string); ok {
							providers = append(providers, pStr)
						}
					}
				}
			}
			if len(providers) != 0 {
				return providers, nil
			}
		}
	}

	return asset.Providers, nil
}

func pushJobs(jobChan chan<- record.OssTask, jobs []record.OssTask) {
	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("push jobs to chan")
	for _, job := range jobs {
		jobChan <- job
	}
}

func (observer *ObserverInfo) worker(id int, jobChan <-chan record.OssTask, ctx context.Context) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	for {
		select {
		case <-ctx.Done():
			workerLogger.Infof("worker-%d stop", id)
			return
		case j := <-jobChan:
			{
				jobId := j.JobId
				traceId := j.TraceId

				jobLogger := log.NewLogicLoggerBuilder().SetTraceId(traceId).SetJobId(jobId).Build()
				jobLogger.Infof("worker-%d start job=%v", id, j)

				//send job request
				err := sendJobRequest(observer.RequestTopic, j)
				if err != nil {
					jobLogger.Error(err.Error())
				}
				jobLogger.Infof("worker-%d sanded job=%v", id, j.JobId)
			}
		}
	}

}

func sendJobRequest(topic string, job record.OssTask) error {

	json, err := json.Marshal(job)
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	eventMsg := PhEventMsg.EventMsg{
		JobId:   job.JobId,
		TraceId: job.TraceId,
		Type:    "PushJob",
		Data:    string(json),
	}

	specificRecordByteArr, err := kafka.EncodeAvroRecord(&eventMsg)
	if err != nil {
		return err
	}

	err = producer.Produce(topic, []byte(job.TraceId), specificRecordByteArr)
	return err
}
